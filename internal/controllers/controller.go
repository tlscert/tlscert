package controllers

import (
	"context"
	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/pkg/errors"
	v1alpha1types "github.com/tlscert/backend/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type CertificatePoolController struct {
	Clients *Clients
}

func NewCertificatePoolController(clients *Clients) *CertificatePoolController {
	return &CertificatePoolController{
		Clients: clients,
	}
}

func (c *CertificatePoolController) Run(ctx context.Context) error {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1types.AddToScheme(scheme))
	utilruntime.Must(certmanager.AddToScheme(scheme))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Metrics: server.Options{
			BindAddress: "0",
		},
		Scheme: scheme, // Set MaxConcurrentReconciles?
	})
	if err != nil {
		return errors.Wrap(err, "failed to create manager")
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1types.CertificatePool{}).
		Owns(&certmanager.Certificate{}).
		Complete(&CertificatePoolReconciler{
			Clients: c.Clients,
		})
	if err != nil {
		return errors.Wrap(err, "failed to create controller")
	}

	klog.Info("starting CertificatePool controller")
	if err = mgr.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start manager")
	}
	return nil
}

type CertificatePoolReconciler struct {
	Clients *Clients
}

func (r *CertificatePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
