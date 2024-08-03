package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"

	"web-server/config_loader"
	"web-server/proxies"

	"github.com/gorilla/mux"
)

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

func NewHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["targetPath"]
		log.Println("Request URL: ", r.URL.String())
		p.ServeHTTP(w, r)
	}
}

func NewAggregatedHandler(targets []config_loader.TargetRoute, versionPorts map[string]int, versionHeader string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		wg.Add(len(targets))

		responses := make([]map[string]interface{}, len(targets))
		for i, target := range targets {
			go func(i int, target config_loader.TargetRoute) {
				defer wg.Done()

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

				resp, err := http.Get(targetURL)
				if err != nil {
					log.Printf("Error fetching from target %v: %v", targetURL, err)
					return
				}
				defer resp.Body.Close()

				var data map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					log.Printf("Error decoding response from target %v: %v", targetURL, err)
					return
				}
				responses[i] = data
			}(i, target)
		}

		wg.Wait()

		aggregatedResponse := make(map[string]interface{})
		for _, resp := range responses {
			for k, v := range resp {
				aggregatedResponse[k] = v
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(aggregatedResponse); err != nil {
			log.Printf("Error encoding aggregated response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func NewChainedHandler(targets []config_loader.TargetRoute, versionPorts map[string]int, versionHeader string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		initialBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(initialBody))

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

			r.Body = ioutil.NopCloser(bytes.NewReader(initialBody)) // Обновляем тело запроса
			httpResp, err := sendRequest(targetURL, r)
			if err != nil {
				log.Printf("Error in chain at target %v: %v", targetURL, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer httpResp.Body.Close()

			finalResponseBody, err = ioutil.ReadAll(httpResp.Body)
			if err != nil {
				log.Printf("Error reading response body from target %v: %v", targetURL, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Обновляем initialBody для следующего запроса в цепочке
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

	// Копируем заголовки из оригинального запроса
	for k, v := range r.Header {
		req.Header[k] = v
	}

	return client.Do(req)
}

// func sendRequest(target string, body []byte) ([]byte, error) {
// 	req, err := http.NewRequest("POST", target, bytes.NewReader(body))
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	return ioutil.ReadAll(resp.Body)
// }
