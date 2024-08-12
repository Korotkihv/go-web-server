package main

import (
	"log"
	"net/http"
	"os"
	"web-server/config_loader"
	"web-server/db"
	"web-server/routes"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL.String(), r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.SetOutput(os.Stdout)

	db.InitGlobalDB("./headerinfo.db")

	cfg, err := config_loader.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Println("Initializing routes...")

	r := routes.InitializeRoutes(cfg)

	loggedRouter := LoggingMiddleware(r)

	log.Printf("Started server on %v", cfg.Gateway.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.Gateway.ListenAddr, loggedRouter))
}
