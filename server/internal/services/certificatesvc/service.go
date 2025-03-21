package certificatesvc

import (
	"context"

	svcpb "github.com/tlscert/tlscert/protos/tlscert/service/v1"
	"github.com/tlscert/tlscert/server/internal/manager"
)

type CertificateService struct {
	svcpb.UnimplementedCertificateServiceServer
	cm *manager.CertificateManager
}

func New(cm *manager.CertificateManager) *CertificateService {
	return &CertificateService{
		cm: cm,
	}
}

func (s *CertificateService) GetCertificate(ctx context.Context, _ *svcpb.GetCertificateRequest) (*svcpb.GetCertificateResponse, error) {
	certificate, err := s.cm.GetCertificate(ctx)
	if err != nil {
		return nil, err
	}

	return &svcpb.GetCertificateResponse{
		Host:        certificate.Host,
		Certificate: [][]byte{certificate.Chain},
		Key:         certificate.Key,
	}, nil
}
