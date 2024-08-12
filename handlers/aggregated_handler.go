package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"web-server/config_loader"
)

func NewAggregatedHandler(targets []config_loader.TargetConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		wg.Add(len(targets))

		responses := make([]map[string]interface{}, len(targets))
		for i, target := range targets {
			go func(i int, target config_loader.TargetConfig) {
				defer wg.Done()

				targetURL, err := target.GetURL(r)
				if err != nil {
					log.Printf("Error get url addr:%v context:%v. %v", target.Addr, target.Context, err)
					return
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
