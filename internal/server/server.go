package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Takes in the desired address to run the server on.
// Returns a HTTPServer pre-configured with the API routes
func NewHTTPServer(addr string) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/ping", handlePing).Methods("GET")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode("pong")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
