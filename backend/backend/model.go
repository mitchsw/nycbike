package backend

import (
	"errors"
	"fmt"
	"strings"
	"time"

	rg "github.com/RedisGraph/redisgraph-go"
	"github.com/gomodule/redigo/redis"
)

// A ModelPool is used to create cheap Model structs used per request.
type ModelPool struct {
	connPool                  redis.Pool
	journeyQueryStringBuilder func(src, dst Circle) string
}

func NewModelPool(address string) (*ModelPool, error) {
	mp := &ModelPool{}
	mp.connPool = redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", address) },
	}
	mp.journeyQueryStringBuilder = journeyQueryStringBuilder()
	return mp, nil
}

func (mp *ModelPool) Close() error {
	return mp.connPool.Close()
}

type Model struct {
	conn                      redis.Conn
	graph                     rg.Graph
	journeyQueryStringBuilder func(src, dst Circle) string
}

// Returns a new Model to be used by a request. Close() should be
// called on the Model before the request ends.
func (mp *ModelPool) Get() *Model {
	m := &Model{}
	m.conn = mp.connPool.Get()
	m.graph = rg.GraphNew("journeys", m.conn)
	m.journeyQueryStringBuilder = mp.journeyQueryStringBuilder
	return m
}

func (m *Model) Close() error {
	return m.conn.Close()
}

type Vitals struct {
	TripCount, StationCount, EdgeCount int
	MemoryUsageHuman                   string
}

func (m *Model) Vitals() (*Vitals, error) {
	var v Vitals
	var err error
	if v.TripCount, err = m.TripCount(); err != nil {
		return nil, err
	}
	if v.StationCount, err = m.StationCount(); err != nil {
		return nil, err
	}
	if v.EdgeCount, err = m.EdgeCount(); err != nil {
		return nil, err
	}
	if v.MemoryUsageHuman, err = m.MemoryUsageHuman(); err != nil {
		return nil, err
	}
	return &v, nil
}

func (m *Model) TripCount() (int, error) {
	return redis.Int(m.conn.Do("GET", "trips"))
}

func (m *Model) StationCount() (int, error) {
	r, err := m.graph.Query("MATCH (s:Station) RETURN count(s)")
	if err != nil {
		return 0, err
	}
	if !r.Next() {
		return 0, nil
	}
	return r.Record().GetByIndex(0).(int), nil
}

func (m *Model) EdgeCount() (int, error) {
	r, err := m.graph.Query("MATCH (:Station)-[t:Trip]->(:Station) RETURN count(t)")
	if err != nil {
		return 0, err
	}
	if !r.Next() {
		return 0, nil
	}
	return r.Record().GetByIndex(0).(int), nil
}

func (m *Model) MemoryUsageHuman() (string, error) {
	info, err := redis.String(m.conn.Do("INFO", "memory"))
	if err != nil {
		return "", err
	}
	for _, l := range strings.Split(info, "\r\n") {
		if strings.HasPrefix(l, "used_memory_human:") {
			return strings.TrimPrefix(l, "used_memory_human:"), nil
		}
	}
	return "", errors.New("cannot find used_memory_human in INFO")
}

type Coord struct {
	Lat, Long float64
}

func (m *Model) GetStations() ([]Coord, error) {
	// WARN: For redisgraph-so to understand RETURNING a point,
	// https://github.com/RedisGraph/redisgraph-go/pull/45 is required.
	res, err := m.graph.Query(
		"MATCH (s:Station) RETURN s.name, s.loc")
	if err != nil {
		return nil, err
	}
	var result []Coord
	for res.Next() {
		r := res.Record()
		pos := r.GetByIndex(1).(map[string]float64)
		result = append(result, Coord{pos["latitude"], pos["longitude"]})
	}
	return result, nil
}

type Circle struct {
	Center   Coord
	RadiusKm float64
}

type JourneyData struct {
	Egress, Ingress []int
	RunTimeMs       float64
}

func journeyQueryStringBuilder() func(src, dst Circle) string {
	// Build a long Graph query which returns a sum every hour in the week.
	// Initially, I used a consise UNWIND query, but in benchmarking this
	// manually-unwound approach was consistently faster.
	var parts strings.Builder
	parts.WriteString(
		`MATCH (src:Station)<-[t:Trip]->(dst:Station)
		 WHERE distance(src.loc, point({latitude: %f, longitude: %f})) < %f
		 AND distance(dst.loc, point({latitude: %f, longitude: %f})) < %f
		 RETURN (startNode(t) = src)`)
	for i := 0; i < (24 * 7); i++ {
		parts.WriteString(fmt.Sprintf(", sum(t.counts[%d])", i))
	}
	queryStringtmpl := parts.String()

	return func(src, dst Circle) string {
		return fmt.Sprintf(
			queryStringtmpl,
			src.Center.Lat, src.Center.Long, src.RadiusKm*1000,
			dst.Center.Lat, dst.Center.Long, dst.RadiusKm*1000)
	}
}

func (m *Model) JourneyQuery(src, dst Circle) (*JourneyData, error) {
	res, err := m.graph.Query(m.journeyQueryStringBuilder(src, dst))
	if err != nil {
		return nil, err
	}
	data := &JourneyData{}
	for res.Next() {
		r := res.Record()
		var counts []int
		for _, v := range r.Values()[1:] {
			// The query's sum(t.count[i]) returns a float for some reason.
			c := int(v.(float64))
			counts = append(counts, c)
		}
		if r.GetByIndex(0).(bool) {
			data.Egress = counts
		} else {
			data.Ingress = counts
		}
	}
	// Sometimes ingress, egress, or both, can be empty.
	if len(data.Egress) == 0 {
		data.Egress = make([]int, 24*7)
	}
	if len(data.Ingress) == 0 {
		data.Ingress = make([]int, 24*7)
	}
	// Returning runtime is helpful to show off performance. :)
	data.RunTimeMs = res.InternalExecutionTime()
	return data, nil
}
