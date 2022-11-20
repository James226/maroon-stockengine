FROM golang:1.19

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./
COPY handlers/health.go handlers/

RUN go build -o /stockengine

CMD "/stockengine"