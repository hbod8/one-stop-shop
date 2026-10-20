package main

import (
	"log"
	"net/http"

	"onestopshop/api"
	"onestopshop/internal/server"
)

func main() {
	srv := server.New()
	mux := http.NewServeMux()
	api.HandlerFromMux(api.NewStrictHandler(srv, nil), mux)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
