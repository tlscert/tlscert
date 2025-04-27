package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	v1 "github.com/tlscert/tlscert/protos/tlscert/service/v1"
	"github.com/tlscert/tlscert/server/internal/kubernetes"
	"github.com/tlscert/tlscert/server/internal/manager"
	"github.com/tlscert/tlscert/server/internal/middleware"
	"github.com/tlscert/tlscert/server/internal/services/certificatesvc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	pool     = flag.String("pool", "manual", "the certificate pool label to use")
	grpcPort = flag.String("grpc-port", "50051", "the port to listen for grpc on")
)

func main() {
	flag.Parse()
	log.Print("Starting tlscert server")

	client, err := kubernetes.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	certificateManager := manager.NewCertificateManager(client, pool)

	// gRPC Server
	svc := certificatesvc.New(certificateManager)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middleware.LoggingInterceptor),
	)
	v1.RegisterCertificateServiceServer(grpcServer, svc)

	errCh := make(chan error)
	// Start gRPC server
	grpcListenPort := net.JoinHostPort("", *grpcPort)
	lis, err := net.Listen("tcp", grpcListenPort)
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

	// Shutdown gRPC server
	grpcServer.GracefulStop()
}
