package controllers

import (
	"errors"
	v1alpha1 "github.com/tlscert/backend/pkg/generated/clientset/versioned"
	"k8s.io/klog/v2"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	certmanagerclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
)

type Clients struct {
	Kubernetes  kubernetes.Interface
	CertManager certmanagerclientset.Interface
	CertPool    v1alpha1.Interface
	Namespace   string
}

// NewClients creates a new Kubernetes client.
// It attempts to use in-cluster config first, then falls back to kubeconfig file.
func NewClients() (*Clients, error) {
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "tlscert" // Fallback for local development
	}

	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	klog.Infof("connecting to kubernetes api at %s", config.Host)

	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	certManager, err := certmanagerclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	certPool, err := v1alpha1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Kubernetes:  k8s,
		CertManager: certManager,
		CertPool:    certPool,
		Namespace:   namespace,
	}, nil
}

// getConfig loads Kubernetes config from in-cluster environment
// or from kubeconfig file
func getConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return nil, errors.New("no kubernetes config found")
}
