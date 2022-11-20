package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	handlers2 "github.com/james226/maroon-stockengine/v2/handlers"
	_ "github.com/microsoft/go-mssqldb"
	"log"
	"net/http"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "log ", log.LstdFlags|log.Ltime|log.Lshortfile)

	db, err := sql.Open("sqlserver", os.Getenv("SQL_CONNECTIONSTRING"))
	if err != nil {
		logger.Fatal("Error creating connection pool: ", err.Error())
	}

	handlers := handlers2.NewHealthHandler(logger, db)
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Handle("/health", handlers)
	logger.Print("Starting HTTP server")

	logger.Fatal(http.ListenAndServe(":10000", myRouter))
}
