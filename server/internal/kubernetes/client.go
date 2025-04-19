package kubernetes

import (
	"errors"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	certmanagerclientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
)

type Client struct {
	Kubernetes  kubernetes.Interface
	CertManager certmanagerclientset.Interface
	Namespace   string
}

// NewClient creates a new Kubernetes client.
// It attempts to use in-cluster config first, then falls back to kubeconfig file.
func NewClient() (*Client, error) {
	namespace, err := getNamespace()
	if err != nil {
		return nil, err
	}

	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	certManager, err := certmanagerclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Kubernetes:  k8s,
		CertManager: certManager,
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

// getNamespace determines the current namespace from environment variables
// or from the Kubernetes service account namespace file.
// Returns "default" if the file doesn't exist.
func getNamespace() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "default", nil
		}
		return "", err
	}

	return string(data), nil
}
