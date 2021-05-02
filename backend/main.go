package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
)

func main() {
	redisAddress := flag.String("redis", "172.18.0.2:6379", "host:port address of Redis")
	log.SetOutput(os.Stdout)
	conn, err := redis.Dial("tcp", *redisAddress)
	if err != nil {
		panic(err)
	}
	n, err := redis.Int(conn.Do("GET", "trips"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %v trips\n", n)
}
