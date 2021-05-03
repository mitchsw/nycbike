package backend

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

type Model struct {
	conn  redis.Conn
	graph rg.Graph
}

func NewModel(address string) (*Model, error) {
	m := &Model{}
	var err error
	m.conn, err = redis.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	m.graph = rg.GraphNew("journeys", m.conn)
	return m, nil
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

type Coord struct {
	Lat, Long float64
}

type Circle struct {
	Center   Coord
	RadiusKm float64
}

type JourneyData struct {
	Egress, Ingress []int
}

func (m *Model) JourneyQuery(src, dst Circle) (*JourneyData, error) {
	// TODO: tidy up massive RETURN statement.
	query := fmt.Sprintf(`
		MATCH (src:Station)<-[t:Trip]->(dst:Station)
		WHERE distance(src.loc, point({latitude: %f, longitude: %f})) < %f
		AND distance(dst.loc, point({latitude: %f, longitude: %f})) < %f
		RETURN (startNode(t) = src), sum(t.counts[0]), sum(t.counts[1]), sum(t.counts[2]), sum(t.counts[3]), sum(t.counts[4]), sum(t.counts[5]), sum(t.counts[6]), sum(t.counts[7]), sum(t.counts[8]), sum(t.counts[9]), sum(t.counts[10]), sum(t.counts[11]), sum(t.counts[12]), sum(t.counts[13]), sum(t.counts[14]), sum(t.counts[15]), sum(t.counts[16]), sum(t.counts[17]), sum(t.counts[18]), sum(t.counts[19]), sum(t.counts[20]), sum(t.counts[21]), sum(t.counts[22]), sum(t.counts[23]), sum(t.counts[24]), sum(t.counts[25]), sum(t.counts[26]), sum(t.counts[27]), sum(t.counts[28]), sum(t.counts[29]), sum(t.counts[30]), sum(t.counts[31]), sum(t.counts[32]), sum(t.counts[33]), sum(t.counts[34]), sum(t.counts[35]), sum(t.counts[36]), sum(t.counts[37]), sum(t.counts[38]), sum(t.counts[39]), sum(t.counts[40]), sum(t.counts[41]), sum(t.counts[42]), sum(t.counts[43]), sum(t.counts[44]), sum(t.counts[45]), sum(t.counts[46]), sum(t.counts[47]), sum(t.counts[48]), sum(t.counts[49]), sum(t.counts[50]), sum(t.counts[51]), sum(t.counts[52]), sum(t.counts[53]), sum(t.counts[54]), sum(t.counts[55]), sum(t.counts[56]), sum(t.counts[57]), sum(t.counts[58]), sum(t.counts[59]), sum(t.counts[60]), sum(t.counts[61]), sum(t.counts[62]), sum(t.counts[63]), sum(t.counts[64]), sum(t.counts[65]), sum(t.counts[66]), sum(t.counts[67]), sum(t.counts[68]), sum(t.counts[69]), sum(t.counts[70]), sum(t.counts[71]), sum(t.counts[72]), sum(t.counts[73]), sum(t.counts[74]), sum(t.counts[75]), sum(t.counts[76]), sum(t.counts[77]), sum(t.counts[78]), sum(t.counts[79]), sum(t.counts[80]), sum(t.counts[81]), sum(t.counts[82]), sum(t.counts[83]), sum(t.counts[84]), sum(t.counts[85]), sum(t.counts[86]), sum(t.counts[87]), sum(t.counts[88]), sum(t.counts[89]), sum(t.counts[90]), sum(t.counts[91]), sum(t.counts[92]), sum(t.counts[93]), sum(t.counts[94]), sum(t.counts[95]), sum(t.counts[96]), sum(t.counts[97]), sum(t.counts[98]), sum(t.counts[99]), sum(t.counts[100]), sum(t.counts[101]), sum(t.counts[102]), sum(t.counts[103]), sum(t.counts[104]), sum(t.counts[105]), sum(t.counts[106]), sum(t.counts[107]), sum(t.counts[108]), sum(t.counts[109]), sum(t.counts[110]), sum(t.counts[111]), sum(t.counts[112]), sum(t.counts[113]), sum(t.counts[114]), sum(t.counts[115]), sum(t.counts[116]), sum(t.counts[117]), sum(t.counts[118]), sum(t.counts[119]), sum(t.counts[120]), sum(t.counts[121]), sum(t.counts[122]), sum(t.counts[123]), sum(t.counts[124]), sum(t.counts[125]), sum(t.counts[126]), sum(t.counts[127]), sum(t.counts[128]), sum(t.counts[129]), sum(t.counts[130]), sum(t.counts[131]), sum(t.counts[132]), sum(t.counts[133]), sum(t.counts[134]), sum(t.counts[135]), sum(t.counts[136]), sum(t.counts[137]), sum(t.counts[138]), sum(t.counts[139]), sum(t.counts[140]), sum(t.counts[141]), sum(t.counts[142]), sum(t.counts[143]), sum(t.counts[144]), sum(t.counts[145]), sum(t.counts[146]), sum(t.counts[147]), sum(t.counts[148]), sum(t.counts[149]), sum(t.counts[150]), sum(t.counts[151]), sum(t.counts[152]), sum(t.counts[153]), sum(t.counts[154]), sum(t.counts[155]), sum(t.counts[156]), sum(t.counts[157]), sum(t.counts[158]), sum(t.counts[159]), sum(t.counts[160]), sum(t.counts[161]), sum(t.counts[162]), sum(t.counts[163]), sum(t.counts[164]), sum(t.counts[165]), sum(t.counts[166]), sum(t.counts[167])
	`, src.Center.Lat, src.Center.Long, src.RadiusKm*1000, dst.Center.Lat, dst.Center.Long, dst.RadiusKm*1000)

	res, err := m.graph.Query(query)
	if err != nil {
		return nil, err
	}
	data := &JourneyData{}
	for res.Next() {
		r := res.Record()
		var counts []int
		for _, v := range r.Values()[1:] {
			// sum(t.count[i]) returns a float for some reason.
			c := int(v.(float64))
			counts = append(counts, c)
		}
		if r.GetByIndex(0).(bool) {
			data.Egress = counts
		} else {
			data.Ingress = counts
		}
	}
	return data, nil
}
