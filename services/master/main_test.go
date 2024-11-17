package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRegisterService(t *testing.T) {
	ms := NewMasterService()
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/services/register", ms.registerService)

	tests := []struct {
		name       string
		service    ServiceInfo
		wantStatus int
	}{
		{
			name: "valid service registration",
			service: ServiceInfo{
				Name: "test-service",
				URL:  "http://localhost:8081",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "duplicate service registration",
			service: ServiceInfo{
				Name: "test-service",
				URL:  "http://localhost:8081",
			},
			wantStatus: http.StatusOK, // Currently allows overwriting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.service)
			req := httptest.NewRequest("POST", "/api/v1/services/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("registerService() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// Verify service was registered
			ms.servicesMux.RLock()
			service, exists := ms.services[tt.service.Name]
			ms.servicesMux.RUnlock()

			if !exists {
				t.Error("service was not registered")
			}
			if service.URL != tt.service.URL {
				t.Errorf("service URL = %v, want %v", service.URL, tt.service.URL)
			}
		})
	}
}

func TestDeregisterService(t *testing.T) {
	ms := NewMasterService()
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/services/deregister", ms.deregisterService)

	// Register a service first
	testService := &ServiceInfo{
		Name: "test-service",
		URL:  "http://localhost:8081",
	}
	ms.services[testService.Name] = testService

	tests := []struct {
		name       string
		service    string
		wantStatus int
	}{
		{
			name:       "valid service deregistration",
			service:    "test-service",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent service",
			service:    "non-existent",
			wantStatus: http.StatusOK, // Currently returns OK even if service doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"serviceName": tt.service})
			req := httptest.NewRequest("POST", "/api/v1/services/deregister", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("deregisterService() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// Verify service was deregistered
			ms.servicesMux.RLock()
			_, exists := ms.services[tt.service]
			ms.servicesMux.RUnlock()

			if exists && tt.name == "valid service deregistration" {
				t.Error("service was not deregistered")
			}
		})
	}
}

func TestGetServicesStatus(t *testing.T) {
	ms := NewMasterService()
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/services/status", ms.getServicesStatus)

	// Register some test services
	testServices := []*ServiceInfo{
		{Name: "service1", URL: "http://localhost:8081", Status: "healthy"},
		{Name: "service2", URL: "http://localhost:8082", Status: "unhealthy"},
	}

	for _, svc := range testServices {
		ms.services[svc.Name] = svc
	}

	req := httptest.NewRequest("GET", "/api/v1/services/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getServicesStatus() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response []ServiceInfo
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != len(testServices) {
		t.Errorf("got %d services, want %d", len(response), len(testServices))
	}

	// Verify each service is present in response
	for _, want := range testServices {
		found := false
		for _, got := range response {
			if got.Name == want.Name {
				found = true
				if got.URL != want.URL || got.Status != want.Status {
					t.Errorf("service %s mismatch: got %+v, want %+v", want.Name, got, want)
				}
				break
			}
		}
		if !found {
			t.Errorf("service %s not found in response", want.Name)
		}
	}
}

func TestProcessInput(t *testing.T) {
	ms := NewMasterService()
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/process-input", ms.processInput)

	tests := []struct {
		name       string
		input      string
		state      map[string]interface{}
		wantStatus int
	}{
		{
			name:  "valid input",
			input: "go north",
			state: map[string]interface{}{
				"location": "start_room",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "empty input",
			input: "",
			state: map[string]interface{}{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]interface{}{
				"userInput":    tt.input,
				"currentState": tt.state,
			})
			req := httptest.NewRequest("POST", "/api/v1/process-input", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("processInput() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if _, ok := response["updatedState"]; !ok {
				t.Error("response missing updatedState field")
			}
			if _, ok := response["actionSummary"]; !ok {
				t.Error("response missing actionSummary field")
			}
		})
	}
}

func TestGetGameState(t *testing.T) {
	ms := NewMasterService()
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/game-state", ms.getGameState)

	// Set up initial game state
	initialState := &GameState{
		CurrentRoom: "start_room",
		Inventory:   []string{"torch", "key"},
		State: map[string]interface{}{
			"health": float64(100),  // Specify as float64 to match JSON unmarshaling
			"score":  float64(0),
		},
	}
	ms.gameState = initialState

	req := httptest.NewRequest("GET", "/api/v1/game-state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getGameState() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response GameState
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify game state matches using reflect.DeepEqual
	if !reflect.DeepEqual(response, *initialState) {
		t.Errorf("got game state = %+v, want %+v", response, initialState)
	}
}
