package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"

	"github.com/smallstep/certinfo"
	svcpb "github.com/tlscert/tlscert/protos/tlscert/service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	address   = flag.String("address", "localhost:50051", "the address to connect to")
	plaintext = flag.Bool("plaintext", false, "use plaintext")
)

func _main() error {
	flag.Parse()
	ctx := context.Background()
	var opts []grpc.DialOption
	if *plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(*address, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := svcpb.NewCertificateServiceClient(conn)
	resp, err := client.GetCertificate(ctx, &svcpb.GetCertificateRequest{})
	if err != nil {
		return err
	}

	for _, cert := range resp.Certificate {
		block, _ := pem.Decode(cert)
		if block == nil {
			return fmt.Errorf("failed to decode certificate")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return err
		}

		certText, err := certinfo.CertificateText(cert)
		if err != nil {
			return err
		}
		fmt.Printf("%s", certText)
	}
	return nil
}

func main() {
	if err := _main(); err != nil {
		log.Fatalf("error: %v", err)
	}
}
