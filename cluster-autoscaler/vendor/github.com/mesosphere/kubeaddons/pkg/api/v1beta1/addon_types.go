package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mesosphere/kubeaddons/pkg/status"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ready",type="string",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="stage",type="string",JSONPath=".status.stage"

// Addon is the Schema for the addons API
type Addon struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata" yaml:"metadata"`

	Spec   AddonSpec   `json:"spec" yaml:"spec"`
	Status AddonStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (a *Addon) GetStatus() *AddonStatus {
	return &a.Status
}

func (a *Addon) GetAddonSpec() *AddonSpec {
	return &a.Spec
}

func (a *Addon) DeployNamespace() string {
	return a.Namespace
}

func (a *Addon) NewEmptyType() AddonInterface {
	return &Addon{}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AddonList contains a list of Addon
type AddonList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Addon `json:"items" yaml:"items"`
}

// AddonSpec defines the desired state of Addon
type AddonSpec struct {
	// ChartReference defines the Helm Chart configuration of this addon (if applicable)
	ChartReference *ChartReference `json:"chartReference,omitempty" yaml:"chartReference,omitempty"`

	// KudoReference defines the KUDO package configuration of this addon (if applicable)
	KudoReference *KudoReference `json:"kudoReference,omitempty" yaml:"kudoReference,omitempty"`

	// Kubernetes defines configuration options relevant to the Kubernetes cluster where the addon will be deployed
	// +optional
	Kubernetes *KubernetesReference `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`

	// Requires (dependencies) based on LabelSelectors. This allows for depending on a specific addon, or
	// a label that defines the capability, ie: test that metadata.labels.cni exists.
	// +optional
	Requires []metav1.LabelSelector `json:"requires,omitempty" yaml:"requires,omitempty"`

	// CloudProvider defines the cloud providers for which this addon should be included in the default configuration.
	// If CloudProvider is omitted, all cloud providers will be included and the default for each will be set to "enabled".
	// If CloudProvider is an empty list, this addon will not be part of a generated list of available addons for
	// implementers of this library to consume. They may still add this addon to their list and enable it.
	// If CloudProvider has any entries, only those cloud providers will receive this addons as an available addon in the
	// generated list. Any cloud providers absent from the list will not receive this addon as "available".
	// TODO: come back and replace "generated list" with the function name that generates this list.
	CloudProvider []ProviderSpec `json:"cloudProvider,omitempty" yaml:"cloudProvider,omitempty"`
}

// AddonStatus defines the observed state of Addon
type AddonStatus struct {
	Ready bool          `json:"ready" yaml:"ready"`
	Stage status.Status `json:"stage,omitempty" yaml:"stage,omitempty"`
}

// ChartReference defines the Helm Chart configuration of this addon (if applicable)
type ChartReference struct {
	// Chart is the name of the desired chart to use
	Chart string `json:"chart" yaml:"chart"`

	// Version is the version of the chart to use
	Version string `json:"version" yaml:"version"`

	// Release is the helm release by which to reference this addon
	// +optional
	Release *string `json:"release,omitempty" yaml:"release,omitempty"`

	// Repo is the chart repository to use (if not the default)
	// +optional
	Repo *string `json:"repo,omitempty" yaml:"repo,omitempty"`

	// Values is the value configurations defined to configure the chart
	// +optional
	Values *string `json:"values,omitempty" yaml:"values,omitempty"`
}

// KudoReference defines the KUDO package configuration of this addon (if applicable)
type KudoReference struct {
	// Package is the name of the desired package to use
	Package string `json:"package" yaml:"package"`

	// Version if the version of the package to use
	Version string `json:"version" yaml:"version"`

	// Repo is the package repository to use
	RepoURL string `json:"repo" yaml:"repo"`

	// Parameters is the parameter configuration defined to configure the package
	// +optional
	Parameters *string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// KubernetesReference defines the kubernetes related configuration options for an addon
type KubernetesReference struct {
	// MinSupportedVersion is the minimum version of Kubernetes that this addon can be used with
	// +optional
	MinSupportedVersion *string `json:"minSupportedVersion,omitempty" yaml:"minSupportedVersion,omitempty"`

	// MaxSupportedVersion is the maximum version of Kubernetes that this addon can be used with
	// +optional
	MaxSupportedVersion *string `json:"maxSupportedVersion,omitempty" yaml:"maxSupportedVersion,omitempty"`
}

// ProviderSpec is configuration specific to a cloud provider
type ProviderSpec struct {
	// Name is the cloud provider name, ie "aws" or "none"
	Name string `json:"name" yaml:"name"`

	// Enabled is a field that can be used by the library implementer to populate a list of available
	// addons. This field will allow them to set a default enable or disable setting.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Values provides provider specific values which should be merged into the general values before deployment
	// +optional
	Values string `json:"values,omitempty" yaml:"values,omitempty"`
}
