package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mitchsw/citibike-journeys/backend/backend"
)

func PrintVitals(m *backend.Model) {
	h := m.Get()
	defer h.Close()

	v, err := h.Vitals()
	if err != nil {
		panic(err)
	}
	log.Printf("Found %v trips, %v stations, %v edges. Memory usage: %v",
		v.TripCount, v.StationCount, v.EdgeCount, v.MemoryUsageHuman)
}

func main() {
	redisAddress := flag.String("redis", "localhost:6379", "host:port address of Redis")
	listenPort := flag.Int("port", 9736, "port to listen on")
	log.SetOutput(os.Stdout)

	m, err := backend.NewModel(*redisAddress)
	if err != nil {
		panic(err)
	}
	log.Println("Connected to Redis!")
	PrintVitals(m)

	a := backend.NewApp(m)
	log.Printf("Running app on port %d...", *listenPort)
	a.Run(fmt.Sprintf(":%d", *listenPort))
}
