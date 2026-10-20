package main

import (
	"fmt"
	"log"
	"net/http"

	"onestopshop/internal/server"
)

var generateSecrets = "false"
var buildTime = "not specified"

func main() {
	fmt.Printf("OneStopShop - Build Time: %s\n", buildTime)
	if generateSecrets == "true" {
		fmt.Println("[WARNING] Generating new secrets on startup. This is not recommended for production use.")
	}
	config, err := server.LoadConfig(buildTime, generateSecrets == "true")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	auth, err := server.NewAuth(config)
	if err != nil {
		log.Fatalf("Failed to initialize auth: %v", err)
	}
	srv := server.New(auth)
	log.Fatal(http.ListenAndServe(":8080", srv.Handler()))
}
