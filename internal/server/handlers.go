package server

import (
	"encoding/json"
	"net/http"

	"github.com/dantdj/wake-on-lan-proxy/internal/esxi"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode("root request received")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	hlog.FromRequest(r).Info().
		Msg("Ping received")

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode("pong")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handlePoweron(w http.ResponseWriter, r *http.Request, ec esxi.Connection) {
	hlog.FromRequest(r).Info().
		Str("MAC Address", ec.MACAddress).
		Msg("Sending Wake on LAN packet")

	if err := ec.TurnOnServer(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error().
			Str("MAC Address", ec.MACAddress).
			Msg("Error occurred when sending Wake on LAN packet")
		return
	}

	hlog.FromRequest(r).Info().
		Str("MAC Address", ec.MACAddress).
		Msg("Wake on LAN packet sent successfully")

	w.WriteHeader(http.StatusOK)
}
