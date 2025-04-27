package manager

import (
	"context"
	"errors"
	"fmt"
	"log"

	"math/rand/v2"

	"github.com/tlscert/tlscert/server/internal"
	"github.com/tlscert/tlscert/server/internal/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CertificateManager struct {
	Client *kubernetes.Client
	Pool   *string
}

func NewCertificateManager(client *kubernetes.Client, pool *string) *CertificateManager {
	return &CertificateManager{
		Client: client,
		Pool:   pool,
	}
}

func (m *CertificateManager) GetCertificate(ctx context.Context) (*internal.Certificate, error) {
	labelSelector := fmt.Sprintf("api.tlscert.dev/pool=%s", *m.Pool)
	log.Printf("Selecting certificates with label %s in namespace %s", labelSelector, m.Client.Namespace)
	certs, err := m.Client.CertManager.CertmanagerV1().Certificates(m.Client.Namespace).List(ctx, v1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil, err
	}

	if len(certs.Items) == 0 {
		return nil, errors.New("no certificate available")
	}

	// TODO: Filter for ready certificates
	// TODO: Mark certificates as used?

	log.Printf("Found %d certificates", len(certs.Items))
	n := len(certs.Items) - 1
	if n > 0 {
		// TODO: something more reasonable
		n = rand.IntN(n)
	}
	cert := certs.Items[n]
	log.Printf("Selected certificate %s", cert.Name)

	secretName := cert.Spec.SecretName

	secret, err := m.Client.Kubernetes.CoreV1().Secrets(m.Client.Namespace).Get(ctx, secretName, v1.GetOptions{})

	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, errors.New("secret data is empty")
	}

	chain := secret.Data["tls.crt"]
	key := secret.Data["tls.key"]

	if chain == nil || key == nil {
		return nil, errors.New("missing certificate data")
	}

	return &internal.Certificate{
		Chain: chain,
		Key:   key,
		Host:  cert.Spec.DNSNames[0],
	}, nil

}
