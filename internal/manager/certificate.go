package manager

import (
	"context"
	"errors"

	"github.com/tlscert/backend/internal"
	"github.com/tlscert/backend/internal/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CertificateManager struct {
	Client *kubernetes.Client
}

func NewCertificateManager(client *kubernetes.Client) *CertificateManager {
	return &CertificateManager{
		Client: client,
	}
}

func (m *CertificateManager) GetCertificate(ctx context.Context) (*internal.Certificate, error) {
	certs, err := m.Client.CertManager.CertmanagerV1().Certificates(m.Client.Namespace).List(ctx, v1.ListOptions{
		LabelSelector: "api.tlscert.dev/pool=manual",
	})

	if err != nil {
		return nil, err
	}

	if len(certs.Items) == 0 {
		return nil, errors.New("no certificate available")
	}

	// TODO: Filter for ready certificates
	// TODO: Mark certificates as used?
	cert := certs.Items[0]

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
