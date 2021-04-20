package importer

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

const citiBikeBucket = "https://s3.amazonaws.com/tripdata/"

// Importer reads trip data from the Citi Bike System Data bucket, and writes
// Trips to the RedisGraph. It is optimised for throughput.
type Importer struct {
	connPool *redis.Pool
	dw       *DataWriter
}

func NewImporter(connPool *redis.Pool, numWorkers, batchSize int) (*Importer, error) {
	dw, err := NewDataWriter(connPool, numWorkers, batchSize)
	if err != nil {
		return nil, err
	}
	return &Importer{connPool: connPool, dw: dw}, nil
}

// Runs the long-running parallel importer. If resetGraph is true, the graph is deleted
// before starting.
func (i *Importer) Run(resetGraph bool) error {
	log.Printf("[importer] Importer running...")
	if resetGraph {
		if err := i.resetGraph(); err != nil {
			return err
		}
	}

	zipFiles, err := scrapeZipFiles()
	if err != nil {
		return err
	}

	conn, err := i.connPool.Dial()
	if err != nil {
		return err
	}

	defer i.dw.Close()
	for idx, zipUrl := range zipFiles {
		resp, err := redis.Int(conn.Do("SISMEMBER", "SCRAPED_FILES", zipUrl))
		if err != nil {
			return err
		}
		if resp > 0 {
			log.Printf("[importer] Already scraped %v", zipUrl)
			continue
		}
		log.Printf("[importer] Scraping %v/%v: %v", idx+1, len(zipFiles), zipUrl)
		if err := i.doImport(zipUrl); err != nil {
			return err
		}
		_, err = conn.Do("SADD", "SCRAPED_FILES", zipUrl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Importer) resetGraph() error {
	log.Printf("[importer] Resetting graph!")
	conn, err := i.connPool.Dial()
	if err != nil {
		return err
	}

	if _, err = redis.Int(conn.Do("DEL", "SCRAPED_FILES")); err != nil {
		return err
	}
	if _, err = redis.Int(conn.Do("DEL", "trips")); err != nil {
		return err
	}

	graph := rg.GraphNew("journeys", conn)
	if err := graph.Delete(); err != nil {
		log.Printf("graph.Delete failed: %v", err)
	}
	if _, err := graph.Query("CREATE INDEX ON :Station(id)"); err != nil {
		return err
	}

	if _, err := graph.Query("CREATE INDEX ON :Station(loc)"); err != nil {
		return err
	}
	return nil
}

func (i *Importer) doImport(zipUrl string) error {
	tdr, err := NewTripdataReader(zipUrl)
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
		i.dw.WriteTrip(t)
	}
	return nil
}

type listObjectsContents struct {
	Key          string
	LastModified time.Time
	Size         int64
}

type listObjectsResp struct {
	Name     string
	Contents []listObjectsContents
}

func scrapeZipFiles() ([]string, error) {
	resp, err := http.Get(citiBikeBucket)
	if err != nil {
		return []string{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("read body: %v", err)
	}

	var objects listObjectsResp
	if err := xml.Unmarshal(data, &objects); err != nil {
		return []string{}, err
	}

	var result []string
	for _, c := range objects.Contents {
		if !strings.HasSuffix(c.Key, ".zip") {
			continue
		}
		result = append(result, citiBikeBucket+c.Key)
	}
	return result, nil
}
