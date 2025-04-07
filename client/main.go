package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
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

	"golang.org/x/sync/errgroup"
)

func _main() error {
	server := flag.String("server", "localhost:8080", "the server to use")
	target := flag.String("target", "http://localhost:8080", "the target port to proxy to")
	port := flag.String("port", "8443", "the port to listen on")
	flag.Parse()

	address := url.URL{Scheme: "http", Host: *server, Path: "/certificate"}

	resp, err := http.Get(address.String()) //nolint:noctx // This is fine. We're probably removing this code anyway.
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	certificateResponse := CertificateResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&certificateResponse); err != nil {
		return err
	}

	pair, err := tls.X509KeyPair(certificateResponse.Chain, certificateResponse.Key)

	if err != nil {
		return err
	}
	listenPort := net.JoinHostPort("", *port)
	targetURL, err := url.Parse(*target)
	if err != nil {
		return err
	}
	log.Printf("Target URL: %s", targetURL)

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
		log.Printf("Starting proxy on https://%s%s", pair.Leaf.DNSNames[0], srv.Addr)
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

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}
