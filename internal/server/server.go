package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
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
			Err(err).
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
			if ec.ServerReachable() && timeDifference.Minutes() > 15 {
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

	proxy, err := NewProxy("https://esxi.dantdj.com:8081/")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/ping", handlePing).Methods("GET")

	r.NewRoute().PathPrefix("/").Methods("GET", "POST", "PUT", "DELETE").HandlerFunc(handleProxy(proxy, ec))
	middlewareRouter := configureMiddleware(r)

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

// Middleware that updates the last request time
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

func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("URL is " + url.String())

	cert, err := tls.LoadX509KeyPair("/etc/letsencrypt/live/dantdj.com/fullchain.pem", "/etc/letsencrypt/live/dantdj.com/privkey.pem")

	proxy := &httputil.ReverseProxy{}
	director := func(req *http.Request) {
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
	}
	proxy.Director = director

	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		},
	}

	return proxy, nil
}
