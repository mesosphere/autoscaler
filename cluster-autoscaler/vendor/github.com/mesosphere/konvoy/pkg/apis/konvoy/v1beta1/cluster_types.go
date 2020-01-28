package v1beta1

import (
	kubeaddons "github.com/mesosphere/kubernetes-resource-manager/pkg/addons"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterConfiguration describes Kubernetes cluster options
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ClusterConfigurationSpec `json:"spec,omitempty"`
}

// ClusterConfigurationSpec is the spec that contains the Kubernetes cluster options
type ClusterConfigurationSpec struct {
	Kubernetes          *Kubernetes          `json:"kubernetes,omitempty"`
	ContainerNetworking *ContainerNetworking `json:"containerNetworking,omitempty"`
	ContainerRuntime    *ContainerRuntime    `json:"containerRuntime,omitempty"`
	ImageRegistries     []ImageRegistry      `json:"imageRegistries,omitempty"`
	OSPackages          *OSPackages          `json:"osPackages,omitempty"`
	NodePools           []NodePool           `json:"nodePools,omitempty"`
	Addons              []Addons             `json:"addons,omitempty"`
	Version             *string              `json:"version,omitempty"`
}

// OSPackages configures the installation of linux package and related properties
type OSPackages struct {
	EnableAdditionalRepositories *bool `json:"enableAdditionalRepositories,omitempty"`
}

// Kubernetes controls the options used by kubeadm and at other points during installation
type Kubernetes struct {
	Version          *string           `json:"version,omitempty"`
	ControlPlane     *ControlPlane     `json:"controlPlane,omitempty"`
	Networking       *Networking       `json:"networking,omitempty"`
	CloudProvider    *CloudProvider    `json:"cloudProvider,omitempty"`
	AdmissionPlugins *AdmissionPlugins `json:"admissionPlugins,omitempty"`
	PreflightChecks  *PreflightChecks  `json:"preflightChecks,omitempty"`
	Kubelet          *Kubelet          `json:"kubelet,omitempty"`
}

// ControlPlane contains all control plane related configurations
type ControlPlane struct {
	ControlPlaneEndpointOverride *string      `json:"controlPlaneEndpointOverride,omitempty"`
	Certificate                  *Certificate `json:"certificate,omitempty"`
	Keepalived                   *Keepalived  `json:"keepalived,omitempty"`
}

// Certificate contains information about an X.509 certificate
type Certificate struct {
	SubjectAlternativeNames []string `json:"subjectAlternativeNames,omitempty"`
}

// Keepalived describes different keepalived related options
type Keepalived struct {
	Interface *string `json:"interface,omitempty"`
	Vrid      *int32  `json:"vrid,omitempty"`
}

// Networking describes different networking related options
type Networking struct {
	PodSubnet     *string  `json:"podSubnet,omitempty"`
	ServiceSubnet *string  `json:"serviceSubnet,omitempty"`
	HTTPProxy     *string  `json:"httpProxy,omitempty"`
	HTTPSProxy    *string  `json:"httpsProxy,omitempty"`
	NoProxy       []string `json:"noProxy,omitempty"`
}

// CloudProvider describes the options passed to Kubernets cloud-provider options
type CloudProvider struct {
	Provider *string `json:"provider,omitempty"`
	//ConfigData *string `json:"configData,omitempty"`
}

type AdmissionPlugins struct {
	Enabled  []string `json:"enabled,omitempty"`
	Disabled []string `json:"disabled,omitempty"`
}

// PreflightChecks describes the set of preflight checks to be performed.
type PreflightChecks struct {
	ErrorsToIgnore []string `json:"errorsToIgnore,omitempty"`
}

// Kubelet describes the settings for the Kubelet.
type Kubelet struct {
	CgroupRoot *string `json:"cgroupRoot,omitempty"`
}

// ContainerNetworking describes the CNI used by Kubernetes
type ContainerNetworking struct {
	Calico *CalicoContainerNetworking `json:"calico,omitempty"`
}

// CalicoContainerNetworking describes Calico CNI
type CalicoContainerNetworking struct {
	Version       *string `json:"version,omitempty"`
	Encapsulation *string `json:"encapsulation,omitempty" yaml:"encapsulation,omitempty"`
	MTU           *int32  `json:"mtu,omitempty"`
}

// ContainerRuntime describes the runtime used by the Kubelet
type ContainerRuntime struct {
	Containerd *ContainerdContainerRuntime `json:"containerd,omitempty"`
}

// ContainerdContainerRuntime describes containerd runtime options
type ContainerdContainerRuntime struct {
	Version    *string     `json:"version,omitempty"`
	ConfigData *ConfigData `json:"configData,omitempty"`
}

// ImageRegistry describes the docker image registries that are automatically configured to be used by the ContainerRuntime
type ImageRegistry struct {
	Server        *string `json:"server,omitempty"`
	Username      *string `json:"username,omitempty"`
	Password      *string `json:"password,omitempty"`
	Auth          *string `json:"auth,omitempty"`
	IdentityToken *string `json:"identityToken,omitempty"`
	Default       *bool   `json:"default,omitempty"`
}

type Addons struct {
	ConfigRepository *string                 `json:"configRepository,omitempry"`
	ConfigVersion    *string                 `json:"configVersion,omitempty"`
	HelmRepository   *HelmRepository         `json:"helmRepository,omitempty"`
	AddonList        kubeaddons.AddonConfigs `json:"addonsList,omitempty"`
}

// NodePool is an object that contains details of a pool such as its name, taints and labels
type NodePool struct {
	Name   string      `json:"name"`
	Labels []NodeLabel `json:"labels,omitempty"`
	Taints []NodeTaint `json:"taints,omitempty"`
	GPU    *GPU        `json:"gpu,omitempty"`
}

// NodeTaint represents a kubernetes taint to be applied to a node
type NodeTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// NodeLabel represents a kubernetes node label
type NodeLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GPU represents an object that contains details of user defined GPU info
type GPU struct {
	Nvidia *Nvidia `json:"nvidia,omitempty"`
}

// Nvidia defines the user configuration of Nvidia specific info
type Nvidia struct{}

// ConfigData represents a file configuration for a Konvoy cluster component, and whether
// or not that should be merged into the corresponding data or completely replace it.
type ConfigData struct {
	Data    string `json:"data"`
	Replace bool   `json:"replace"`
}

type HelmRepository struct {
	Image *string `json:"image,omitempty"`
}
