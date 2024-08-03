package routes

import (
	"log"
	"web-server/config_loader"
	"web-server/handlers"
	"web-server/proxies"

	"github.com/gorilla/mux"
)

func InitializeRoutes(cfg *config_loader.Config) *mux.Router {
	r := mux.NewRouter()

	initializeBasicRoutes(r, cfg)
	initializeAggregatedRoutes(r, cfg)
	initializeChainedRoutes(r, cfg)

	return r
}

func initializeBasicRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.Routes {
		if route.Versioned {
			// TODO make abstract
			r.HandleFunc(route.Context, handlers.NewVersionedHandler(route.Target, cfg.Gateway.VersionPorts, cfg.Gateway.VersionHeader))
		} else {
			proxy, err := proxies.NewProxy(route.Target)
			if err != nil {
				panic(err)
			}
			log.Printf("Mapping '%v' | %v ---> %v", route.Name, route.Context, route.Target)
			r.HandleFunc(route.Context, handlers.NewHandler(proxy))
		}
	}
}

func initializeAggregatedRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.AggregatedRoutes {
		targets := make([]config_loader.TargetRoute, len(route.Targets))
		copy(targets, route.Targets)

		log.Printf("Mapping aggregated route '%v' | %v ---> %v", route.Name, route.Context, targets)

		r.HandleFunc(route.Context, handlers.NewAggregatedHandler(targets, cfg.Gateway.VersionPorts, cfg.Gateway.VersionHeader))
	}
}

func initializeChainedRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.ChainedRoutes {
		targets := make([]config_loader.TargetRoute, len(route.Targets))
		copy(targets, route.Targets)

		log.Printf("Mapping chained route '%v' | %v ---> %v", route.Name, route.Context, targets)

		r.HandleFunc(route.Context, handlers.NewChainedHandler(targets, cfg.Gateway.VersionPorts, cfg.Gateway.VersionHeader))
	}
}
