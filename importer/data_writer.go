package importer

import (
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

type DataWriter struct {
	connPool   *redis.Pool
	numWorkers int
	batchSize  int

	// Optimisation: cache the stations we have already created.
	stationsCreated sync.Map

	trips   chan Trip
	done    chan bool
	workers []dataWriterWorker
}

type dataWriterWorker struct {
	id          int
	dw          *DataWriter
	conn        redis.Conn
	pipelineCnt int
	tripCnt     int
}

func NewDataWriter(pool *redis.Pool, numWorkers, batchSize int) (*DataWriter, error) {
	dw := &DataWriter{
		connPool:   pool,
		numWorkers: numWorkers,
		batchSize:  batchSize,
		trips:      make(chan Trip),
		done:       make(chan bool),
	}
	for i := 1; i <= dw.numWorkers; i++ {
		dw.addWorker()
		go dw.workers[len(dw.workers)-1].Run()
	}
	return dw, nil
}

func (dw *DataWriter) Close() {
	close(dw.trips)
	for i := 1; i <= dw.numWorkers; i++ {
		<-dw.done
	}
}

func (dw *DataWriter) WriteTripdata(zip_file string) error {
	tdr, err := NewTripdataReader(zip_file)
	if err != nil {
		return err
	}
	defer tdr.Close()
	for {
		t, err := tdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		dw.trips <- *t
	}
	return nil
}

func (dw *DataWriter) addWorker() {
	id := len(dw.workers)
	dw.workers = append(dw.workers, dataWriterWorker{
		id: id,
		dw: dw,
	})
}

func (dww *dataWriterWorker) Run() {
	var err error
	dww.conn, err = dww.dw.connPool.Dial()
	if err != nil {
		panic(err)
	}
	//dww.conn = redis.NewLoggingConn(dww.conn, log.Default(), fmt.Sprintf("[dww.%v.redis]", dww.id))
	log.Printf("[dww.%v]: Started", dww.id)

	err = dww.Send("CLIENT", "REPLY", "OFF")
	if err != nil {
		panic(err)
	}

	for t := range dww.dw.trips {
		if err := dww.importTrip(&t); err != nil {
			panic(err)
		}
	}

	// Final flush
	if err := dww.FlushPipeline(); err != nil {
		panic(err)
	}
	if err := dww.Send("INCRBY", "trips", dww.tripCnt); err != nil {
		panic(err)
	}

	log.Printf("[dataWriterWorker.%v]: Done", dww.id)

	dww.dw.done <- true
}

func (dww *dataWriterWorker) importTrip(t *Trip) error {
	err := dww.maybeCreateStation(t.StartStationId, t.StartStationName, t.StartStationLat, t.StartStationLong)
	if err != nil {
		return err
	}
	err = dww.maybeCreateStation(t.EndStationId, t.EndStationName, t.EndStationLat, t.EndStationLong)
	if err != nil {
		return err
	}

	err = dww.addTripEdge(t.StartStationId, t.EndStationId, t.StartTime)
	if err != nil {
		return err
	}

	dww.tripCnt++
	if dww.tripCnt > 200 {
		err = dww.Send("INCRBY", "trips", dww.tripCnt)
		if err != nil {
			return err
		}
		dww.tripCnt = 0
	}

	return nil
}

func (dww *dataWriterWorker) addTripEdge(startStationId, endStationId int, t time.Time) error {
	day := int(t.Weekday())
	hour := day*24 + t.Hour()
	//log.Printf("(s:Station{id: %v})-[e:Trips{h: %v}]->(s:Station{id: %v})", startStationId, hour, endStationId)
	q := fmt.Sprintf(`
		MATCH (src:Station{id: $src})
		MATCH (dst:Station{id: $dst})
		MERGE (src)-[t:Trips{h:%d}]->(dst)
		ON CREATE SET t.count = 1
		ON MATCH SET t.count = t.count+1
	`, hour)
	return dww.SendGraphQuery(q, map[string]interface{}{
		"src": startStationId, "dst": endStationId,
	})
}

func (dww *dataWriterWorker) maybeCreateStation(id int, name string, lat, long float64) error {
	if _, ok := dww.dw.stationsCreated.Load(id); ok {
		return nil
	}
	q := `
	OPTIONAL MATCH (s:Station{id: $id})
	WITH COUNT(s) AS c WHERE c = 0
	CREATE (:Station{
		id: $id,
		name: $name,
		loc: point({latitude: $lat, longitude: $long})
	})
	`
	err := dww.SendGraphQuery(q, map[string]interface{}{
		"id": id, "name": name, "lat": lat, "long": long,
	})
	if err == nil {
		dww.dw.stationsCreated.Store(id, true)
	}
	return err
}

func (dww *dataWriterWorker) SendGraphQuery(q string, params map[string]interface{}) error {
	err := dww.conn.Send("GRAPH.QUERY", "journeys", rg.BuildParamsHeader(params)+q, "--compact")
	if err != nil {
		return err
	}
	dww.pipelineCnt++
	if dww.pipelineCnt >= dww.dw.batchSize {
		if err := dww.FlushPipeline(); err != nil {
			return err
		}
	}
	return nil
}

func (dww *dataWriterWorker) Send(commandName string, args ...interface{}) error {
	err := dww.conn.Send(commandName, args...)
	if err != nil {
		return err
	}
	dww.pipelineCnt++
	if dww.pipelineCnt >= dww.dw.batchSize {
		if err := dww.FlushPipeline(); err != nil {
			return err
		}
	}
	return nil
}

func (dww *dataWriterWorker) FlushPipeline() error {
	log.Printf("[dataWriterWorker.%v]: Flushing %v commands", dww.id, dww.pipelineCnt)
	if err := dww.conn.Flush(); err != nil {
		return err
	}
	dww.pipelineCnt = 0
	return nil
}
