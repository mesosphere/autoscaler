package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mesosphere/kubeaddons/pkg/constants"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ready",type="string",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="stage",type="string",JSONPath=".status.stage"

// ClusterAddon is the Schema for the cluster-scoped addons API
type ClusterAddon struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata" yaml:"metadata"`

	Spec   ClusterAddonSpec `json:"spec" yaml:"spec"`
	Status AddonStatus      `json:"status,omitempty" yaml:"status,omitempty"`
}

func (a *ClusterAddon) GetStatus() *AddonStatus {
	return &a.Status
}

func (a *ClusterAddon) DeployNamespace() string {
	if a.Spec.Namespace != nil {
		return *a.Spec.Namespace
	}
	return constants.DefaultAddonNamespace
}

func (a *ClusterAddon) GetAddonSpec() *AddonSpec {
	return &a.Spec.AddonSpec
}

func (a *ClusterAddon) SetAddonStatus(status *AddonStatus) {
	a.Status = *status
}

func (a *ClusterAddon) NewEmptyType() AddonInterface {
	return &ClusterAddon{}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClusterAddonList contains a list of Addon
type ClusterAddonList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []ClusterAddon `json:"items" yaml:"items"`
}

// ClusterAddonSpec defines the desired state of Addon
type ClusterAddonSpec struct {
	// Namespace defines the namespace for which to deploy cluster-scoped addon components to (defaults to the same namespace where the cluster-scoped addon is installed)
	// +optional
	Namespace *string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	AddonSpec `json:",inline" yaml:",inline"`
}
