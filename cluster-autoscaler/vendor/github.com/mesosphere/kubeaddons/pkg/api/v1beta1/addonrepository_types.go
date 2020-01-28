package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="ready",type="string",JSONPath=".status.ready"

// AddonRepository is the Schema for the addonrepositories API
type AddonRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonRepositorySpec   `json:"spec,omitempty"`
	Status AddonRepositoryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AddonRepositoryList contains a list of AddonRepository
type AddonRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonRepository `json:"items"`
}

// AddonRepositorySpec defines the desired state of AddonRepository
type AddonRepositorySpec struct {
	Priority resource.Quantity      `json:"priority,omitempty"`
	URL      string                 `json:"url"`
	Ref      string                 `json:"ref"`
	Type     string                 `json:"type,omitempty"`
	Options  AddonRepositoryOptions `json:"options,omitempty"`
}

// AddonRepositoryOptions defines the credentials reference for the repository
type AddonRepositoryOptions struct {
	CredentialsRef corev1.LocalObjectReference `json:"credentialsRef"`
}

// AddonRepositoryStatus defines the observed state of AddonRepository
type AddonRepositoryStatus struct {
	Ready bool `json:"ready" yaml:"ready"`
}
