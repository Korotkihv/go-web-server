package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

func NewHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["targetPath"]

		log.Println("Request URL: ", r.URL.String())

		p.ServeHTTP(w, r)
	}
}
