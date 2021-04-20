package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/mitchsw/citibike-journies/importer"
)

func main() {
	log.SetOutput(os.Stdout)
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redismod-lb-984e5d74c26f3026.elb.us-east-1.amazonaws.com:6379")
		},
	}
	defer pool.Close()

	imp, err := importer.NewImporter(pool, 100, 250)
	if err != nil {
		panic(err)
	}
	err = imp.Run( /*resetGraph=*/ false)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}
