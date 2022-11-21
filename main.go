package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/james226/stockengine/handlers"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	logger := log.New(os.Stdout, "log ", log.LstdFlags|log.Ltime|log.Lshortfile)

	db, err := sql.Open("sqlserver", os.Getenv("SQL_CONNECTIONSTRING"))
	if err != nil {
		logger.Fatal("Error creating connection pool: ", err.Error())
	}

	handler := handlers.NewHealthHandler(logger, db)
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Handle("/health", handler)
	logger.Print("Starting HTTP server")

	logger.Fatal(http.ListenAndServe(":10000", myRouter))
}
