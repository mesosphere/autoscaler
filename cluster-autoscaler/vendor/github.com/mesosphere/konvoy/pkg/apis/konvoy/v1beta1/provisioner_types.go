package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterProvisioner describes provisioner options
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterProvisioner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ClusterProvisionerSpec `json:"spec,omitempty"`
}

// ClusterProvisionerSpec is the spec that contains the provisioner options
type ClusterProvisionerSpec struct {
	Provider              string                 `json:"provider"`
	AWSProviderOptions    *AWSProviderOptions    `json:"aws,omitempty"`
	AzureProviderOptions  *AzureProviderOptions  `json:"azure,omitempty"`
	DockerProviderOptions *DockerProviderOptions `json:"docker,omitempty"`
	NodePools             []MachinePool          `json:"nodePools,omitempty"`
	SSHCredentials        *SSHCredentials        `json:"sshCredentials,omitempty"`
	Version               *string                `json:"version,omitempty"`
}

// MachinePool used by the provisioner to configure a machine
type MachinePool struct {
	Name         string   `json:"name"`
	ControlPlane *bool    `json:"controlPlane,omitempty"`
	Bastion      *bool    `json:"bastion,omitempty"`
	Count        int32    `json:"count"`
	Machine      *Machine `json:"machine,omitempty"`
}

// Machine used by the provisioner to configure a machine
type Machine struct {
	ImageID              *string           `json:"imageID,omitempty"`
	ImageName            *string           `json:"imageName,omitempty"`
	RootVolumeSize       *int64            `json:"rootVolumeSize,omitempty"`
	RootVolumeType       *string           `json:"rootVolumeType,omitempty"`
	RootVolumeIOPS       *int32            `json:"rootVolumeIOPS,omitempty"`
	ImagefsVolumeEnabled *bool             `json:"imagefsVolumeEnabled,omitempty"`
	ImagefsVolumeSize    *int64            `json:"imagefsVolumeSize,omitempty"`
	ImagefsVolumeType    *string           `json:"imagefsVolumeType,omitempty"`
	ImagefsVolumeDevice  *string           `json:"imagefsVolumeDevice,omitempty"`
	Type                 *string           `json:"type,omitempty"`
	AWSMachineOpts       *AWSMachineOpts   `json:"aws,omitempty"`
	AzureMachineOpts     *AzureMachineOpts `json:"azure,omitempty"`
}

// SSHCredentials describes the options passed to the provisioner regarding the ssh keys
type SSHCredentials struct {
	User           string  `json:"user"`
	PublicKeyFile  string  `json:"publicKeyFile"`
	PrivateKeyFile *string `json:"privateKeyFile,omitempty"`
}

// DockerProviderOptions describes Docker provider related options
type DockerProviderOptions struct {
	DisablePortMapping            *bool  `json:"disablePortMapping,omitempty"`
	ControlPlaneMappedPortBase    *int32 `json:"controlPlaneMappedPortBase,omitempty"`
	SSHControlPlaneMappedPortBase *int32 `json:"sshControlPlaneMappedPortBase,omitempty"`
	SSHWorkerMappedPortBase       *int32 `json:"sshWorkerMappedPortBase,omitempty"`
	DedicatedNetwork              *bool  `json:"dedicatedNetwork,omitempty"`
}
