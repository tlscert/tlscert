package main

import (
	"context"
	"errors"
	"github.com/tlscert/backend/internal/api"
	"github.com/tlscert/backend/internal/kubernetes"
	"github.com/tlscert/backend/internal/manager"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	client, err := kubernetes.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	certificateManager := manager.NewCertificateManager(client)

	server := api.NewServer(certificateManager)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: server,
	}

	go func() {
		log.Printf("Starting server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown gracefully
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server shutdown complete")
}
