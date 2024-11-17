package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// ServiceInfo represents a registered service
type ServiceInfo struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	LastCheck time.Time `json:"lastCheck"`
}

// GameState represents the current state of the game
type GameState struct {
	CurrentRoom string                 `json:"currentRoom"`
	Inventory   []string              `json:"inventory"`
	State       map[string]interface{} `json:"state"`
}

// MasterService is the main orchestrator
type MasterService struct {
	services    map[string]*ServiceInfo
	gameState   *GameState
	servicesMux sync.RWMutex
	stateMux    sync.RWMutex
}

func NewMasterService() *MasterService {
	return &MasterService{
		services:  make(map[string]*ServiceInfo),
		gameState: &GameState{
			State: make(map[string]interface{}),
		},
	}
}

// registerService handles service registration
func (ms *MasterService) registerService(w http.ResponseWriter, r *http.Request) {
	var service ServiceInfo
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ms.servicesMux.Lock()
	ms.services[service.Name] = &service
	ms.servicesMux.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Service registered successfully",
	})
}

// deregisterService handles service deregistration
func (ms *MasterService) deregisterService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServiceName string `json:"serviceName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ms.servicesMux.Lock()
	delete(ms.services, req.ServiceName)
	ms.servicesMux.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Service deregistered successfully",
	})
}

// getServicesStatus returns the status of all registered services
func (ms *MasterService) getServicesStatus(w http.ResponseWriter, r *http.Request) {
	ms.servicesMux.RLock()
	defer ms.servicesMux.RUnlock()

	services := make([]ServiceInfo, 0, len(ms.services))
	for _, svc := range ms.services {
		services = append(services, *svc)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)
}

// processInput handles user input processing via Ollama
func (ms *MasterService) processInput(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserInput    string                 `json:"userInput"`
		CurrentState map[string]interface{} `json:"currentState"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement Ollama integration
	// For now, return a mock response
	response := map[string]interface{}{
		"updatedState": req.CurrentState,
		"actionSummary": "Command processed successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getGameState returns the current game state
func (ms *MasterService) getGameState(w http.ResponseWriter, r *http.Request) {
	ms.stateMux.RLock()
	defer ms.stateMux.RUnlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ms.gameState)
}

// healthCheck periodically checks the health of registered services
func (ms *MasterService) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ms.servicesMux.Lock()
			for name, svc := range ms.services {
				// Perform health check
				resp, err := http.Get(svc.URL + "/health")
				if err != nil {
					svc.Status = "unhealthy"
				} else {
					resp.Body.Close()
					svc.Status = "healthy"
				}
				svc.LastCheck = time.Now()
				ms.services[name] = svc
			}
			ms.servicesMux.Unlock()
		}
	}
}

func main() {
	ms := NewMasterService()
	router := mux.NewRouter()

	// Service management endpoints
	router.HandleFunc("/api/v1/services/register", ms.registerService).Methods("POST")
	router.HandleFunc("/api/v1/services/deregister", ms.deregisterService).Methods("POST")
	router.HandleFunc("/api/v1/services/status", ms.getServicesStatus).Methods("GET")

	// Game management endpoints
	router.HandleFunc("/api/v1/process-input", ms.processInput).Methods("POST")
	router.HandleFunc("/api/v1/game-state", ms.getGameState).Methods("GET")

	// Start health check routine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ms.healthCheck(ctx)

	port := os.Getenv("MASTER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Master Service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
