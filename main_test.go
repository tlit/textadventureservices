package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterService(t *testing.T) {
	service := Service{Name: "TestService", URL: "http://localhost:8081", Status: "active"}
	body, _ := json.Marshal(service)

	req, err := http.NewRequest("POST", "/api/v1/services/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(registerService)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Service registered successfully"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestDeregisterService(t *testing.T) {
	service := Service{Name: "TestService", URL: "http://localhost:8081", Status: "active"}
	body, _ := json.Marshal(service)

	req, err := http.NewRequest("POST", "/api/v1/services/deregister", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(deregisterService)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Service deregistered successfully"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetServiceStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/services/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getServiceStatus)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var services []Service
	if err := json.NewDecoder(rr.Body).Decode(&services); err != nil {
		t.Errorf("handler returned invalid JSON: %v", err)
	}
}
