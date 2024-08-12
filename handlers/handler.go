package handlers

import (
	"log"
	"net/http"
	"web-server/config_loader"
	"web-server/proxies"

	"github.com/gorilla/mux"
)

func NewHandler(target *config_loader.TargetConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["targetPath"]

		target, err := target.GetURL(r)
		log.Printf("KEKLOL3 %s", target)
		if err != nil {
			panic(err)
		}

		proxy, err := proxies.NewProxy(target)
		if err != nil {
			panic(err)
		}

		log.Println("Request URL: ", r.URL.String())
		proxy.ServeHTTP(w, r)
	}
}

func GroupHandler(target *config_loader.TargetConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		target, err := target.GetURL(r)
		log.Printf("KEKLOL3 %s", target)
		if err != nil {
			panic(err)
		}

		proxy, err := proxies.NewProxy(target)
		if err != nil {
			panic(err)
		}

		log.Println("Request URL: ", r.URL.String())
		proxy.ServeHTTP(w, r)
	}
}
