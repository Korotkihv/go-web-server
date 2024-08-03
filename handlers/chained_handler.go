package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"web-server/config_loader"
)

func NewChainedHandler(targets []config_loader.TargetRoute, versionPorts map[string]int, versionHeader string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		initialBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(initialBody))

		var finalResponseBody []byte

		for _, target := range targets {
			targetURL := target.Target
			if target.Versioned {
				version := r.Header.Get(versionHeader)
				port, ok := versionPorts[version]
				if !ok {
					http.Error(w, "Version not supported", http.StatusBadRequest)
					return
				}
				targetURL = fmt.Sprintf("http://host.docker.internal:%d%s", port, target.Target)
			}

			log.Println("Sending request to:", targetURL)
			log.Println("Request body:", string(initialBody))

			r.Body = io.NopCloser(bytes.NewReader(initialBody))
			httpResp, err := sendRequest(targetURL, r)
			if err != nil {
				log.Printf("Error in chain at target %v: %v", targetURL, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer httpResp.Body.Close()

			finalResponseBody, err = io.ReadAll(httpResp.Body)
			if err != nil {
				log.Printf("Error reading response body from target %v: %v", targetURL, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			initialBody = finalResponseBody
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(finalResponseBody)
	}
}

func sendRequest(url string, r *http.Request) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		return nil, err
	}

	for k, v := range r.Header {
		req.Header[k] = v
	}

	return client.Do(req)
}
