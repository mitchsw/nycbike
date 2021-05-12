module github.com/mitchsw/citibike-journeys/backend

go 1.16

require (
	github.com/RedisGraph/redisgraph-go v1.0.1-0.20210228181239-852ca0de336f
	github.com/gomodule/redigo v1.8.4
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/olekukonko/tablewriter v0.0.5 // indirect
)

// My fork includes https://github.com/RedisGraph/redisgraph-go/pull/45.
// DELETE ME and update the main version once merged.
replace github.com/RedisGraph/redisgraph-go => github.com/mitchsw/redisgraph-go v0.2.0
