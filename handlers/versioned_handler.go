package handlers

import (
	"fmt"
	"log"
	"net/http"

	"web-server/proxies"

	"github.com/gorilla/mux"
)

// TODO
func NewVersionedHandler(target string, versionPorts map[string]int, versionHeader string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get(versionHeader)
		port, ok := versionPorts[version]
		if !ok {
			http.Error(w, "Version not supported", http.StatusBadRequest)
			return
		}
		targetURL := fmt.Sprintf("http://host.docker.internal:%d%s", port, target)

		r.URL.Path = mux.Vars(r)["targetPath"]
		proxy, err := proxies.NewProxy(targetURL)
		if err != nil {
			log.Printf("Error creating proxy for target %v: %v", targetURL, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		proxy.ServeHTTP(w, r)
	}
}
