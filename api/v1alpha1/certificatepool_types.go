package v1alpha1

import (
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CertificatePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificatePool `json:"items"`
}

// Represents a CertificatePool CRD.
// +kubebuilder:object:root=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type CertificatePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CertificatePoolSpec `json:"spec"`

	//+kubebuilder:validation:Optional
	//+optional
	Status CertificatePoolStatus `json:"status"`
}

// TODO: Doc
type CertificatePoolSpec struct {

	// +kubebuilder:validation:Required
	CertificateTemplate *v1.CertificateSpec `json:"certificateTemplate"`

	// +kubebuilder:validation:Required
	MinCertificates int `json:"minCertificates"`
	// +kubebuilder:validation:Required
	MaxCertificates int `json:"maxCertificates"`
}

type CertificatePoolStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}
