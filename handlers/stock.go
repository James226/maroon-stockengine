package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type StockHandler struct {
	logger *log.Logger
	db     *sql.DB
}

type ItemStockResponse struct {
	Item     string  `json:"item"`
	Location string  `json:"location"`
	Quantity float64 `json:"quantity"`
	Uom      string  `json:"uom"`
}

func NewStockHandler(logger *log.Logger, db *sql.DB) *StockHandler {
	return &StockHandler{
		logger: logger,
		db:     db,
	}
}

func (h *StockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item := vars["item"]
	location := vars["location"]

	rows, err := h.db.QueryContext(
		r.Context(),
		"SELECT [Item], [Location], [FinalStock], [UOM] FROM [Stock] WHERE [Item] = @item AND [Location] = @location",
		sql.Named("item", item),
		sql.Named("location", location),
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	if !rows.Next() {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	response := ItemStockResponse{}

	err = rows.Scan(&response.Item, &response.Location, &response.Quantity, &response.Uom)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
