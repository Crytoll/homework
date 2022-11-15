FROM golang:1.17-alpine as builder

ENV CGO_ENABLED=0 GO111MODULE=on GOPROXY=https://goproxy.cn,direct

COPY httpserver /app

WORKDIR /app

RUN GOOS=linux go build -installsuffix cgo -o httpserver main.go

FROM alpine:3.15

WORKDIR /app

COPY --from=builder /app/httpserver /app/
COPY httpserver/conf/config.yaml /app/conf/

USER 1000

CMD ["/app/httpserver"]
