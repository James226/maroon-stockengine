package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/james226/stockengine/handlers"
	"github.com/james226/stockengine/migrations"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	logger := log.New(os.Stdout, "log ", log.LstdFlags|log.Ltime|log.Lshortfile)

	connectionString := os.Getenv("SQL_CONNECTIONSTRING")
	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		logger.Fatal("Error creating connection pool: ", err.Error())
	}

	err = migrations.Migrate(db, 1669062011)
	if err != nil {
		logger.Fatal(err)
	}

	healthHandler := handlers.NewHealthHandler(logger, db)
	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/health", healthHandler)

	stockHandler := handlers.NewStockHandler(logger, db)
	router.Handle("/location/{location}/stock/{item}", stockHandler)

	logger.Print("Starting HTTP server")

	logger.Fatal(http.ListenAndServe(":10000", router))
}
