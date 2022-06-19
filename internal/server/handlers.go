package server

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/dantdj/wake-on-lan-proxy/internal/esxi"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// ProxyRequestHandler handles the http request using proxy
func handleProxy(proxy *httputil.ReverseProxy, ec esxi.Connection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ec.ServerReachable() {
			if err := ec.TurnOnServer(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Error().
					Str("MAC Address", ec.MACAddress).
					Err(err).
					Msg("Error occurred when sending Wake on LAN packet")
				return
			}
			log.Info().Msg("Server successfully turned on, waiting 60 seconds before serving requests")
			time.Sleep(60 * time.Second)
		}
		proxy.ServeHTTP(w, r)
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
