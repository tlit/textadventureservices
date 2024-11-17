package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"textadventureservices/services/logging"
)

func main() {
	// noob: Make our quantum logger
	logger := logging.NewQuantumLogger(
		logging.WithBufferSize(4096), // bigger buffer for more quantum entanglement
	)
	defer logger.Shutdown()

	// noob: Set up our HTTP handler
	handler := logging.NewLoggingHandler(logger)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// noob: Start the server in a quantum superposition
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// noob: Handle graceful shutdown through quantum tunneling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		server.Close()
	}()

	// noob: Start our quantum server
	log.Printf("Quantum Logger listening on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Quantum collapse: %v", err)
	}
}
