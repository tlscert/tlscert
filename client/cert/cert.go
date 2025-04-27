package cert

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"slices"

	svcpb "github.com/tlscert/tlscert/protos/tlscert/service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Options struct {
	Endpoint  string
	Plaintext bool
}

type FetchCertificateResponse struct {
	Certificates   []*x509.Certificate
	CertficatePEMs [][]byte

	PrivateKey    crypto.PrivateKey
	PrivateKeyPEM []byte

	Host string
}

func FetchCertificate(fetchCertificateOptions Options) (*FetchCertificateResponse, error) {
	ctx := context.Background()
	var opts []grpc.DialOption
	if fetchCertificateOptions.Plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(fetchCertificateOptions.Endpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := svcpb.NewCertificateServiceClient(conn)
	resp, err := client.GetCertificate(ctx, &svcpb.GetCertificateRequest{})
	if err != nil {
		return nil, err
	}

	ret := &FetchCertificateResponse{
		Host:           resp.Host,
		CertficatePEMs: slices.Clone(resp.Certificate),
	}

	for _, cert := range resp.Certificate {
		block, _ := pem.Decode(cert)
		if block == nil {
			return nil, fmt.Errorf("failed to decode certificate")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		ret.Certificates = append(ret.Certificates, cert)
	}

	ret.PrivateKeyPEM = resp.Key
	block, _ := pem.Decode(resp.Key)
	// We'll probably need to change this later.
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ret.PrivateKey = key

	return ret, nil
}
