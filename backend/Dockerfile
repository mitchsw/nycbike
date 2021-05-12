FROM alpine AS base
RUN apk add --no-cache curl wget

FROM golang:1.11 AS go-builder
WORKDIR /go/app
COPY . /go/app
RUN GO111MODULE=on  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/app/main /go/app/main.go

FROM base
COPY --from=go-builder /go/app/main /main
ENTRYPOINT ["/main"]
CMD ["--redis=localhost:6379", "--port=80"]
