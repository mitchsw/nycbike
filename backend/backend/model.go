package backend

import (
	"errors"
	"fmt"
	"strings"
	"time"

	rg "github.com/RedisGraph/redisgraph-go"
	"github.com/gomodule/redigo/redis"
)

type Model struct {
	connPool                  redis.Pool
	journeyQueryStringBuilder func(src, dst Circle) string
}

func NewModel(address string) (*Model, error) {
	m := &Model{}
	m.connPool = redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", address) },
	}
	m.journeyQueryStringBuilder = journeyQueryStringBuilder()
	return m, nil
}

type ModelHandle struct {
	conn                      redis.Conn
	graph                     rg.Graph
	journeyQueryStringBuilder func(src, dst Circle) string
}

// Returns a new ModelHandle using the internal redis connection pool.
// ModelHandle.Close() should be called before destroying the handle.
func (m *Model) Get() *ModelHandle {
	h := &ModelHandle{}
	h.conn = m.connPool.Get()
	h.graph = rg.GraphNew("journeys", h.conn)
	h.journeyQueryStringBuilder = m.journeyQueryStringBuilder
	return h
}

func (h *ModelHandle) Close() error {
	return h.conn.Close()
}

type Vitals struct {
	TripCount, StationCount, EdgeCount int
	MemoryUsageHuman                   string
}

func (h *ModelHandle) Vitals() (*Vitals, error) {
	var v Vitals
	var err error
	if v.TripCount, err = h.TripCount(); err != nil {
		return nil, err
	}
	if v.StationCount, err = h.StationCount(); err != nil {
		return nil, err
	}
	if v.EdgeCount, err = h.EdgeCount(); err != nil {
		return nil, err
	}
	if v.MemoryUsageHuman, err = h.MemoryUsageHuman(); err != nil {
		return nil, err
	}
	return &v, nil
}

func (h *ModelHandle) TripCount() (int, error) {
	return redis.Int(h.conn.Do("GET", "trips"))
}

func (h *ModelHandle) StationCount() (int, error) {
	r, err := h.graph.Query("MATCH (s:Station) RETURN count(s)")
	if err != nil {
		return 0, err
	}
	if !r.Next() {
		return 0, nil
	}
	return r.Record().GetByIndex(0).(int), nil
}

func (h *ModelHandle) EdgeCount() (int, error) {
	r, err := h.graph.Query("MATCH (:Station)-[t:Trip]->(:Station) RETURN count(t)")
	if err != nil {
		return 0, err
	}
	if !r.Next() {
		return 0, nil
	}
	return r.Record().GetByIndex(0).(int), nil
}

func (h *ModelHandle) MemoryUsageHuman() (string, error) {
	info, err := redis.String(h.conn.Do("INFO", "memory"))
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

func (h *ModelHandle) GetStations() ([]Coord, error) {
	res, err := h.graph.Query(
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

func (h *ModelHandle) JourneyQuery(src, dst Circle) (*JourneyData, error) {
	res, err := h.graph.Query(h.journeyQueryStringBuilder(src, dst))
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
