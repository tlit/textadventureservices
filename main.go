package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Service struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

var services = make(map[string]Service)
var mu sync.Mutex

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/services/register", registerService)
	mux.HandleFunc("/api/v1/services/deregister", deregisterService)
	mux.HandleFunc("/api/v1/services/status", getServiceStatus)

	// Import UI service handlers
	registerUIHandlers(mux)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func registerService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	services[service.Name] = service
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service registered successfully"))
}

func deregisterService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	delete(services, service.Name)
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service deregistered successfully"))
}

func getServiceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var serviceList []Service
	for _, service := range services {
		serviceList = append(serviceList, service)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serviceList)
}
