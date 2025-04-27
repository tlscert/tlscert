package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	svcpb "github.com/tlscert/tlscert/protos/tlscert/service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"golang.org/x/sync/errgroup"
)

var (
	address   = flag.String("address", "localhost:50051", "the address to connect to")
	plaintext = flag.Bool("plaintext", false, "use plaintext")
	target    = flag.String("target", "http://localhost:8080", "the target port to proxy to")
	port      = flag.String("port", "8443", "the port to listen on")
)

func _main() error {
	flag.Parse()

	pair, err := getCertificatePair()
	if err != nil {
		return err
	}
	listenPort := net.JoinHostPort("", *port)
	targetURL, err := url.Parse(*target)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	srv := &http.Server{
		Addr:    listenPort,
		Handler: proxy,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
		},
	}

	wg := errgroup.Group{}

	wg.Go(func() error {
		log.Printf("Starting local proxy on https://%s%s to %s", pair.Leaf.DNSNames[0], srv.Addr, targetURL)
		if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
		return nil
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown gracefully
	log.Println("Shutting down proxy...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("proxy shutdown error: %w", err)
	}

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for proxy to shutdown: %w", err)
	}

	log.Println("proxy shutdown complete")
	return nil
}

func getCertificatePair() (tls.Certificate, error) {
	ctx := context.Background()
	var opts []grpc.DialOption
	if *plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	conn, err := grpc.NewClient(*address, opts...)
	if err != nil {
		return tls.Certificate{}, err
	}
	defer conn.Close()

	client := svcpb.NewCertificateServiceClient(conn)
	resp, err := client.GetCertificate(ctx, &svcpb.GetCertificateRequest{})
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := make([]byte, 0)
	for _, b := range resp.Certificate {
		certPEM = append(certPEM, b...)
	}

	return tls.X509KeyPair(certPEM, resp.Key)
}

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}
