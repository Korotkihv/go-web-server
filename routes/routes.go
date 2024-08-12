package routes

import (
	"log"
	"web-server/config_loader"
	"web-server/handlers"

	"github.com/gorilla/mux"
)

func InitializeRoutes(cfg *config_loader.Config) *mux.Router {
	r := mux.NewRouter()

	initializeBasicRoutes(r, cfg)
	initializeGroupRoutes(r, cfg)
	initializeAggregatedRoutes(r, cfg)
	initializeChainedRoutes(r, cfg)

	return r
}

func initializeBasicRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.Routes {
		log.Printf("Mapping '%v' | %v ---> %v", route.Name, route.Context, route.Target)
		r.HandleFunc(route.Context, handlers.NewHandler(&route.Target))
	}
}

func initializeGroupRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.GroupeRoutes {
		log.Printf("Mapping group: '%v' | %v ---> %v", route.Name, route.ContextPrefix, route.Target)

		r.PathPrefix(route.ContextPrefix).HandlerFunc(handlers.GroupHandler(&route.Target))
	}
}

func initializeAggregatedRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.AggregatedRoutes {
		targets := make([]config_loader.TargetConfig, len(route.Targets))

		copy(targets, route.Targets)

		log.Printf("Mapping aggregated route '%v' | %v ---> %v", route.Name, route.Context, targets)

		r.HandleFunc(route.Context, handlers.NewAggregatedHandler(targets))
	}
}

func initializeChainedRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.ChainedRoutes {
		targets := make([]config_loader.TargetConfig, len(route.Targets))

		copy(targets, route.Targets)

		log.Printf("Mapping chained route '%v' | %v ---> %v", route.Name, route.Context, targets)

		r.HandleFunc(route.Context, handlers.NewChainedHandler(targets))
	}
}

// func initializeWebSocketRoutes(r *mux.Router, cfg *config_loader.Config) {
// 	// for _, route := range cfg.Gateway.WebSocketRoutes {
// 	// 	log.Printf("Mapping WebSocket route '%v' | %v ---> %v", route.Name, route.Context, route.Target)
// 	// 	playerHubRoutes := r.PathPrefix(route.Context).Subrouter()
// 	// 	playerHubRoutes.PathPrefix("/").HandlerFunc(handlers.WebSocketProxy(route.Target))
// 	// }
// }
