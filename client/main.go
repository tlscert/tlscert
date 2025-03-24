package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := flag.String("server", "localhost:8080", "the server to use")
	target := flag.String("target", "http://localhost:8080", "the target port to proxy to")
	port := flag.String("port", "8443", "the port to listen on")
	flag.Parse()

	address := url.URL{Scheme: "http", Host: *server, Path: "/certificate"}

	resp, err := http.Get(address.String())
	if err != nil {
		log.Fatal(err)
	}

	certificateResponse := CertificateResponse{}
	json.NewDecoder(resp.Body).Decode(&certificateResponse)

	pair, err := tls.X509KeyPair(certificateResponse.Chain, certificateResponse.Key)

	if err != nil {
		log.Fatal(err)
	}
	listenPort := net.JoinHostPort("", *port)
	targetUrl, err := url.Parse(*target)
	log.Printf("Target URL: %s", targetUrl)

	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	srv := &http.Server{
		Addr:    listenPort,
		Handler: proxy,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
		},
	}

	go func() {
		log.Printf("Starting proxy on https://%s%s", pair.Leaf.DNSNames[0], srv.Addr)
		if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start proxy: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown gracefully
	log.Println("Shutting down proxy...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err2 := srv.Shutdown(ctx); err2 != nil {
		log.Printf("proxy shutdown error: %v", err2)
	}

	log.Println("proxy shutdown complete")
}
