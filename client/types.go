package main

type CertificateResponse struct {
	Chain []byte `json:"chain"`
	Key   []byte `json:"key"`
	Host  string `json:"host"`
}
