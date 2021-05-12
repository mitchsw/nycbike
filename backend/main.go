package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mitchsw/citibike-journeys/backend/backend"
)

func PrintVitals(mp *backend.ModelPool) {
	m := mp.Get()
	defer m.Close()

	v, err := m.Vitals()
	if err != nil {
		panic(err)
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
