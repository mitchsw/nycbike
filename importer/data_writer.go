package importer

import (
	"log"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

// DataWriter writes Trips to the RedisGraph. It is optimised for throughput,
// using concurrent writers, each writing batches of CREATE commands.
type DataWriter struct {
	connPool   *redis.Pool
	numWorkers int
	batchSize  int

	// Optimisation: cache the station IDs we have already created.
	stationsCreated sync.Map

	trips   chan *Trip
	done    chan bool
	workers []dataWriterWorker
}

type dataWriterWorker struct {
	id   int
	dw   *DataWriter
	conn redis.Conn

	pipelineCnt int // The number of commands waiting to be flushed.
	tripCnt     int // The number of trips written.
}

func NewDataWriter(pool *redis.Pool, numWorkers, batchSize int) (*DataWriter, error) {
	dw := &DataWriter{
		connPool:   pool,
		numWorkers: numWorkers,
		batchSize:  batchSize,
		trips:      make(chan *Trip),
		done:       make(chan bool),
	}
	for i := 1; i <= dw.numWorkers; i++ {
		dw.addWorker()
		go dw.workers[len(dw.workers)-1].Run()
	}
	return dw, nil
}

// Calling Close is important to flush any final batches of Trips.
func (dw *DataWriter) Close() {
	close(dw.trips)
	for i := 1; i <= dw.numWorkers; i++ {
		<-dw.done
	}
}

// Asynchronously writes a Trip to the graph. Failures will panic.
func (dw *DataWriter) WriteTrip(t *Trip) {
	dw.trips <- t
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

	if err := dww.conn.Send("CLIENT", "REPLY", "OFF"); err != nil {
		panic(err)
	}

	// Main consumer loop
	for t := range dww.dw.trips {
		if err := dww.writeTrip(t); err != nil {
			panic(err)
		}
	}

	// Final flush
	if err := dww.flushPipeline(); err != nil {
		panic(err)
	}

	if err := dww.conn.Close(); err != nil {
		panic(err)
	}
	dww.dw.done <- true
	log.Printf("[dww.%v]: Done", dww.id)
}

func (dww *dataWriterWorker) writeTrip(t *Trip) error {
	dww.tripCnt++
	err := dww.maybeCreateStation(t.StartStationId, t.StartStationName, t.StartStationLat, t.StartStationLong)
	if err != nil {
		return err
	}
	err = dww.maybeCreateStation(t.EndStationId, t.EndStationName, t.EndStationLat, t.EndStationLong)
	if err != nil {
		return err
	}
	return dww.addTripEdge(t.StartStationId, t.EndStationId, t.StartTime)
}

func (dww *dataWriterWorker) addTripEdge(startStationId, endStationId int, t time.Time) error {
	hour := int(t.Weekday())*24 + t.Hour()
	q := `
		MATCH (src:Station{id: $src})
		MATCH (dst:Station{id: $dst})
		MERGE (src)-[t:Trip]->(dst)
		ON CREATE SET t.counts = [n in range(0, 167) | CASE WHEN n = $hour THEN 1 ELSE 0 END] 
		ON MATCH SET t.counts = t.counts[0..$hour] + [t.counts[$hour]+1] + t.counts[($hour+1)..168]
	`
	return dww.SendGraphQuery(q, map[string]interface{}{
		"src": startStationId, "dst": endStationId, "hour": hour,
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
	return dww.Send("GRAPH.QUERY", "journeys", rg.BuildParamsHeader(params)+q, "--compact")
}

func (dww *dataWriterWorker) Send(commandName string, args ...interface{}) error {
	err := dww.conn.Send(commandName, args...)
	if err != nil {
		return err
	}
	dww.pipelineCnt++
	if dww.pipelineCnt >= dww.dw.batchSize {
		if err = dww.flushPipeline(); err != nil {
			return err
		}
	}
	return nil
}

func (dww *dataWriterWorker) flushPipeline() error {
	log.Printf("[dataWriterWorker.%v]: Flushing %v commands, %v trips", dww.id, dww.pipelineCnt, dww.tripCnt)
	if err := dww.conn.Send("INCRBY", "trips", dww.tripCnt); err != nil {
		return err
	}
	// We toggle CLIENT REPLY ON then OFF so the server can signal the pipeline batch
	// has been consumed.
	if err := dww.conn.Send("CLIENT", "REPLY", "ON"); err != nil {
		return err
	}
	if err := dww.conn.Send("CLIENT", "REPLY", "OFF"); err != nil {
		return err
	}
	if err := dww.conn.Flush(); err != nil {
		return err
	}
	// Block until the CLIENY TRPLU ON -> OK response is consumed.
	if _, err := dww.conn.Receive(); err != nil {
		return err
	}

	dww.pipelineCnt = 0
	dww.tripCnt = 0
	return nil
}
