package server

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dantdj/wake-on-lan-proxy/internal/esxi"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	EsxiConnection esxi.Connection
	Server         http.Server
}

// TODO: This should get wrapped in a mutex at some point so
// concurrent requests don't cause undefined behaviour
var lastRequestTime time.Time

// Takes in the desired address to run the server on.
// Returns a HTTPServer pre-configured with the API routes, and containing the request timer
func New(addr string) *Server {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().
			Msg("Failed to load .env file")
	}

	ec := esxi.Connection{
		Username:   os.Getenv("ESXI_USER"),
		Password:   os.Getenv("ESXI_PASS"),
		URL:        os.Getenv("ESXI_URL"),
		MACAddress: os.Getenv("ESXI_MAC"),
	}

	// Kick off goroutine to track the time since last request
	lastRequestTime = time.Now()
	go func() {
		for {
			timeDifference := time.Now().Sub(lastRequestTime)
			if ec.ServerReachable() && timeDifference.Seconds() > 15 {
				log.Info().
					Time("lastRequestTime", lastRequestTime).
					Time("currentTime", time.Now()).
					Float64("timeDifference", timeDifference.Seconds()).
					Msg("Deadline since last request time exceeded, sending power-off request")

				ec.TurnOffServer()
			}

			time.Sleep(1 * time.Minute)
		}
	}()

	r := mux.NewRouter()
	middlewareRouter := configureMiddleware(r)

	r.HandleFunc("/", handleRoot).Methods("GET")
	r.HandleFunc("/ping", handlePing).Methods("GET")
	r.HandleFunc("/poweron", func(w http.ResponseWriter, r *http.Request) {
		handlePoweron(w, r, ec)
	}).Methods("GET")

	server := http.Server{
		Addr:    addr,
		Handler: middlewareRouter,
	}

	return &Server{
		EsxiConnection: ec,
		Server:         server,
	}
}

func configureMiddleware(r *mux.Router) http.Handler {
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	chain := alice.New()

	// Add logging configuration
	chain = chain.Append(hlog.NewHandler(log))

	// Log each request to the server
	chain = chain.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))

	// Add fields to append to all log messages
	chain = chain.Append(
		hlog.RemoteAddrHandler("ip"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("req_id", "Request-Id"),
	)

	// Add request tracking
	chain = chain.Append(recordRequest)

	return chain.Then(r)
}

// Middleware that updates the last request time to track when to disable the server
func recordRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip ping requests, it only matters if a poweron or traffic request has come through
		if !strings.Contains(r.RequestURI, "/ping") {
			previousLastRequestTime := lastRequestTime
			lastRequestTime = time.Now()

			hlog.FromRequest(r).Info().
				Time("previousLastRequestTime", previousLastRequestTime).
				Time("newLastRequestTime", lastRequestTime).
				Msg("updated lastRequestTime")
		}

		next.ServeHTTP(w, r)
	})
}
