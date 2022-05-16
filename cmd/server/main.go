package main

import (
	"log"

	"github.com/dantdj/wake-on-lan-proxy/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
