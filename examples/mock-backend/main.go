// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Kyber-Proxy") != "true" {
			log.Printf("[REJECTED] %s from %s", r.URL.Path, r.RemoteAddr)
			http.Error(w, "Unauthorized: must go through KyberLink Gateway", http.StatusUnauthorized)
			return
		}

		log.Printf("[REQUEST] %s", r.URL.Path)

		var payload interface{}
		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				log.Printf("[ERROR] Decode: %v", err)
			}
		}

		log.Printf("[DATA] %+v", payload)

		response := map[string]interface{}{
			"status":   "success",
			"endpoint": r.URL.Path,
			"msg":      "Payload processed by Final Server",
			"received": payload,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	http.HandleFunc("/test1", handler)
	http.HandleFunc("/test2", handler)
	http.HandleFunc("/test3", handler)

	log.Println("Mock backend on :34890 (requires X-Kyber-Proxy)")
	log.Fatal(http.ListenAndServe(":34890", nil))
}
