package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mitchsw/citibike-journeys/backend/backend"
)

func PrintVitals(mp *backend.ModelPool) {
	m := mp.Get()
	defer m.Close()

	// On start up, keep polling on LOADING errors.
	v, err := m.Vitals()
	for err != nil {
		if !strings.HasPrefix(err.Error(), "LOADING") {
			panic(err)
		}
		log.Println(err)
		time.Sleep(5 * time.Second)
		v, err = m.Vitals()
	}

	log.Printf("Found %v trips, %v stations, %v edges. Memory usage: %v",
		v.TripCount, v.StationCount, v.EdgeCount, v.MemoryUsageHuman)
}

func main() {
	redisAddress := flag.String("redis", "localhost:6379", "host:port address of Redis")
	listenPort := flag.Int("port", 80, "port to listen on")
	flag.Parse()

	log.SetOutput(os.Stdout)

	mp, err := backend.NewModelPool(*redisAddress)
	if err != nil {
		panic(err)
	}
	log.Println("Connected to Redis!")
	PrintVitals(mp)

	a := backend.NewApp(mp)
	log.Printf("Running app on port %d...", *listenPort)
	a.Run(fmt.Sprintf(":%d", *listenPort))
}
