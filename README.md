# Citibike Journeys

_[Build on Redis Hackathon](https://redislabs.com/hackathon-2021/) entry, mitchsw, 2021-05-12._

A visual geospatial index of over 56 million bikeshare trips across NYC. This could be helpful to capacity plan across the network, allowing you to investigate aggregated rush hour and weekend travel patterns in milliseconds!

**Live Demo**: https://citibike.mitchsw.com/

<img src="https://raw.githubusercontent.com/mitchsw/citibike-journeys/main/full_ui.png?raw=true">*Full visual UI.*

<img src="https://raw.githubusercontent.com/mitchsw/citibike-journeys/main/zoom_ui.png?raw=true">*Zoomed-in view of trips between a few stations.*

## System Overview

The visual UI is built using:
  1.  [RedisGraph](https://oss.redislabs.com/redisgraph/) through [redismod](https://hub.docker.com/r/redislabs/redismod),
  2. a Go backend (behind an nginx reverse proxy),
  3. a React frontend. 

This infrastructure can be started from docker-compose.yml.

This repo also includes a Go importer program to load the public dataset into RedisGraph.

### redismod

This project uses the [redismod](https://hub.docker.com/r/redislabs/redismod) Docker image. This was used (as per Hackathon requirements) instead Redis Enterprise Cloud as that did not yet support RedisGraph v2.4 (at time of development).

### backend

The Go backend uses the [redisgraph-go](https://github.com/RedisGraph/redisgraph-go) library to proxy graph queries from the frontend. The Go library didn't support the new `point()` type, so I sent RedisGraph/redisgraph-go#45 adding this feature.

To mark every station on the map (`/stations` API call), a simple Cypher query is used:

```sql
MATCH (s:Station) RETURN count(s)
```

To count all the edges in the graph (part of `/vitals` API call), another simple Cypher query is used:

```sql
MATCH (:Station)-[t:Trip]->(:Station) RETURN count(t)
```

The main Cypher query to retrieve journeys (`/journey_query` API call) is of the form:

```sql
MATCH (src:Station)<-[t:Trip]->(dst:Station)
WHERE distance(src.loc, point($src)) < $src_radius
  AND distance(dst.loc, point($dst)) < $dst_radius
RETURN
  (startNode(t) = src) as egress,
  sum(t.counts[0]) as h0_trip_count,
  ...
```

This matches all the `:Stations` within the `$src` and `$dst` circles, and all the trip edges between these stations in both directions. This is a fast query due to the geospatial index on `:Station.loc` (see _offline_importer_ below). The returned `egress` is true if the trip started at `$src`, or false if it started at `$dst`. The aggregated trip graph presented on the UI is built by aggregating properties on these `:Trip` edges, for both egress and ingress traffic.

### frontend

The frontend is built in React, built around [react-mapbox-gl](https://github.com/alex3165/react-mapbox-gl) and custom drawing modes I implemented. The aggregated trip graph is built using [devexpress/dx-react-chart](https://github.com/DevExpress/devextreme-reactive).

This is my _first ever_ React project, be nice! ;)

### offline_importer

The offline importer iteratively downloads the public [Citi Bike trip data](https://www.citibikenyc.com/system-data), unzips each archive, and indexes all the trips into the `journeys` graph.

The graph contains every `:Station` as a node, and a [geospatial index](https://oss.redislabs.com/redisgraph/commands/#indexing) of their location. All of the 56 million journeys are represented as increments on the edge between the `src` and `dst` stations. The graph is setup to aggregate trips based on their time of the week (`7*24` hour buckets). This graph could easily be extended to aggregate trips on other dimensions.

To index a single trip, the following Cypher query is used:

```sql
MATCH (src:Station{id: $src})
MATCH (dst:Station{id: $dst})
MERGE (src)-[t:Trip]->(dst)
ON CREATE
  SET t.counts = [n in range(0, 167) | CASE WHEN n = $hour THEN 1 ELSE 0 END] 
ON MATCH
  SET t.counts = t.counts[0..$hour] + [t.counts[$hour]+1] + t.counts[($hour+1)..168]
```

This either created the edge, or increments the appropriate counter to index the trip.

To efficiently write all 56 million trips, I use [pipelining](https://redis.io/topics/pipelining) and [`CLIENT REPLY OFF`](https://redis.io/commands/client-reply) for each batch. The bulk import takes a couple of hours.

## How to run

Start the visual UI using Docker Compose:

```sh
 docker build -t citibike-journeys backend
 docker-compose up
```

Then, start indexing the public dataset:

```sh
$ go run main.go --reset_graph=true
2021/05/12 22:58:45 [importer] Importer running...
2021/05/12 22:58:45 [importer] Resetting graph!
2021/05/12 22:58:45 [dww.0]: Started
2021/05/12 22:58:46 [importer] Scraping 1/164: https://s3.amazonaws.com/tripdata/201306-citibike-tripdata.zip
2021/05/12 22:58:47 [tripdata_reader] Opened file: 201306-citibike-tripdata.csv
2021/05/12 22:58:47 [dataWriterWorker.0]: Flushing 10000 commands, 9668 trips
2021/05/12 22:58:52 [dataWriterWorker.0]: Flushing 10000 commands, 9998 trips
2021/05/12 22:58:56 [dataWriterWorker.0]: Flushing 10000 commands, 10000 trips
2021/05/12 22:59:01 [dataWriterWorker.0]: Flushing 10000 commands, 10000 trips
2021/05/12 22:59:05 [dataWriterWorker.0]: Flushing 10000 commands, 10000 trips
```

Each reload of the UI at http://localhost/ should show these trips accumulate.