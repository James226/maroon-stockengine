FROM golang:1.19

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./
COPY handlers/health.go handlers/
COPY handlers/stock.go handlers/
COPY migrations/migrate.go migrations/
COPY migrations/sqlserver.go migrations/
COPY migrations/sql/1669062011_create_stock_table.down.sql migrations/sql/
COPY migrations/sql/1669062011_create_stock_table.up.sql migrations/sql/

RUN go build -o /stockengine

CMD "/stockengine"