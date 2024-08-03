package main

import (
	"log"
	"net/http"
	"os"
	"web-server/config_loader"
	"web-server/routes"

	"github.com/gorilla/mux"
)

func main() {
	log.SetOutput(os.Stdout)

	cfg, err := config_loader.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Println("Initializing routes...")

	r := mux.NewRouter()

	routes.InitializeRoutes(r, cfg)

	log.Printf("Started server on %v", cfg.Gateway.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.Gateway.ListenAddr, r))
}
