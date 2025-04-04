package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/tlscert/tlscert/protos/tlscert/service/v1"
	"github.com/tlscert/tlscert/server/internal/api"
	"github.com/tlscert/tlscert/server/internal/kubernetes"
	"github.com/tlscert/tlscert/server/internal/manager"
	"github.com/tlscert/tlscert/server/internal/middleware"
	"github.com/tlscert/tlscert/server/internal/services/certificatesvc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	client, err := kubernetes.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	certificateManager := manager.NewCertificateManager(client)

	// HTTP Server
	server := api.NewServer(certificateManager)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: server,
	}

	// gRPC Server
	svc := certificatesvc.New(certificateManager)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middleware.LoggingInterceptor),
	)
	v1.RegisterCertificateServiceServer(grpcServer, svc)

	errCh := make(chan error)
	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("Starting gRPC server on %s", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()
	// Register reflection service on gRPC server
	reflection.Register(grpcServer)
	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigCh)

	// Wait until signal is received
	select {
	case err := <-errCh:
		log.Printf("Server error: %v", err)
	case <-sigCh:
		log.Println("Signal received, shutting down servers...")
	}

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()
}
