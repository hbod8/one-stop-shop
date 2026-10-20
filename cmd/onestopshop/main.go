package main

import (
	"log"
	"net/http"

	"onestopshop/api"
	"onestopshop/internal/server"
)

func main() {
	config, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	auth := server.NewAuth(config)
	srv := server.New(auth)
	mux := http.NewServeMux()
	api.HandlerFromMux(api.NewStrictHandler(srv, nil), mux)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
