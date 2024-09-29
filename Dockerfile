FROM golang:1.23-bullseye as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" \
    --mount=type=cache,target="/root/.config/go" \
    go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.io,direct \
    && go mod download
COPY . .

RUN go build -o collector_server .

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/collector_server .

CMD ["./collector_server"]
