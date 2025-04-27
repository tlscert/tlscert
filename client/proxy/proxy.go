package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"golang.org/x/sync/errgroup"
)

type CreateProxyOptions struct {
	ListenPort    string
	TargetAddress string

	Keypair tls.Certificate
}

func CreateProxy(ctx context.Context, opts CreateProxyOptions) error {
	listenPort := net.JoinHostPort("", opts.ListenPort)
	targetURL, err := url.Parse(opts.TargetAddress)
	if err != nil {
		return err
	}
	log.Printf("Target URL: %s", targetURL)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	srv := &http.Server{
		Addr:    listenPort,
		Handler: proxy,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{opts.Keypair},
		},
	}

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		log.Printf("Starting proxy on https://%s%s", opts.Keypair.Leaf.DNSNames[0], srv.Addr)
		if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Println("Shutting down server...")
		return srv.Close()
	})

	return wg.Wait()
}
