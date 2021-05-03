package main

import (
	"flag"
	"log"
	"os"

	"github.com/mitchsw/citibike-journeys/backend/backend"
)

func printStartupStatus(m *backend.Model) {
	log.Println("Connected to Redis!")
	tc, err := m.TripCount()
	if err != nil {
		panic(err)
	}
	sc, err := m.StationCount()
	if err != nil {
		panic(err)
	}
	ec, err := m.EdgeCount()
	if err != nil {
		panic(err)
	}
	log.Printf("Found %v trips, %v stations, %v edges\n", tc, sc, ec)
}

func main() {
	redisAddress := flag.String("redis", "172.18.0.2:6379", "host:port address of Redis")
	log.SetOutput(os.Stdout)

	m, err := backend.NewModel(*redisAddress)
	if err != nil {
		panic(err)
	}
	printStartupStatus(m)

	a := backend.NewApp(m)
	log.Println("Running app..")
	a.Run(":80")
}
