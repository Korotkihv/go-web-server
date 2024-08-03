package routes

import (
	"log"
	"web-server/config_loader"
	"web-server/handlers"
	"web-server/proxies"

	"github.com/gorilla/mux"
)

func InitializeRoutes(r *mux.Router, cfg *config_loader.Config) {
	for _, route := range cfg.Gateway.Routes {
		if route.Versioned {
			r.HandleFunc(route.Context+"/{targetPath:.*}", handlers.NewVersionedHandler(route.Target, cfg.Gateway.VersionPorts, cfg.Gateway.VersionHeader))
		} else {
			proxy, err := proxies.NewProxy(route.Target)
			if err != nil {
				panic(err)
			}
			log.Printf("Mapping '%v' | %v ---> %v", route.Name, route.Context, route.Target)
			r.HandleFunc(route.Context, handlers.NewHandler(proxy))
		}
	}

	for _, route := range cfg.Gateway.AggregatedRoutes {
		targets := make([]config_loader.TargetRoute, len(route.Targets))
		for i, target := range route.Targets {
			targets[i] = target
		}
		log.Printf("Mapping aggregated route '%v' | %v ---> %v", route.Name, route.Context, targets)
		r.HandleFunc(route.Context, handlers.NewAggregatedHandler(targets, cfg.Gateway.VersionPorts, cfg.Gateway.VersionHeader))
	}

	for _, route := range cfg.Gateway.ChainedRoutes {
		targets := make([]config_loader.TargetRoute, len(route.Targets))
		for i, target := range route.Targets {
			targets[i] = target
		}
		log.Printf("Mapping chained route '%v' | %v ---> %v", route.Name, route.Context, targets)
		r.HandleFunc(route.Context, handlers.NewChainedHandler(targets, cfg.Gateway.VersionPorts, cfg.Gateway.VersionHeader))
	}
}
