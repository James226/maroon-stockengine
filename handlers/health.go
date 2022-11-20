package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

type HealthHandler struct {
	logger *log.Logger
	db     *sql.DB
}

func NewHealthHandler(logger *log.Logger, db *sql.DB) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		db:     db,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.db.PingContext(r.Context())
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Healthy"))
}
