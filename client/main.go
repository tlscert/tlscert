package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/tlscert/tlscert/client/cert"
	"github.com/tlscert/tlscert/client/proxy"
)

var (
	serverEndpoint = flag.String("server", "localhost:50051", "the address to connect to")
	port           = flag.String("port", "8443", "the port to listen on")
	target         = flag.String("target", "http://localhost:8080", "the target host to proxy to")
	plaintext      = flag.Bool("plaintext", false, "use plaintext")
)

func _main() error {
	flag.Parse()

	resp, err := cert.FetchCertificate(cert.Options{
		Plaintext: *plaintext,
		Endpoint:  *serverEndpoint,
	})
	if err != nil {
		return err
	}

	pair, err := tls.X509KeyPair(resp.CertficatePEMs[0], resp.PrivateKeyPEM)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	return proxy.CreateProxy(ctx, proxy.CreateProxyOptions{
		ListenPort:    *port,
		TargetAddress: *target,
		Keypair:       pair,
	})
}

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}
