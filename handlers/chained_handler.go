package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"web-server/config_loader"
)

func NewChainedHandler(targets []config_loader.TargetConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		initialBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(initialBody))

		var finalResponseBody []byte

		for _, target := range targets {
			targetURL, err := target.GetURL(r)
			if err != nil {
				log.Printf("Error get url addr:%v context:%v. %v", target.Addr, target.Context, err)
				return
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
