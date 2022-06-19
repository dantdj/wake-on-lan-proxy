package main

import (
	"github.com/dantdj/wake-on-lan-proxy/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	srv := server.New(":8080")

	log.Info().
		Msg("Starting server on port 8080")

	log.Fatal().Err(srv.Server.ListenAndServe()).Msg("Error received")
}
