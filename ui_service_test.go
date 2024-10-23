package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetUIState(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/ui/state", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getUIState)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var actual map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&actual); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	expected := map[string]interface{}{"state": "active"}
	if !equalJSON(actual, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", actual, expected)
	}
}

// Helper function to compare two JSON objects
func equalJSON(a, b map[string]interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func TestPostUICommand(t *testing.T) {
	command := Command{Command: "move north"}
	body, _ := json.Marshal(command)

	req, err := http.NewRequest("POST", "/api/v1/ui/command", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(postUICommand)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var actual map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&actual); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	expected := map[string]interface{}{"message": "Command processed successfully"}
	if !equalJSON(actual, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", actual, expected)
	}
}
