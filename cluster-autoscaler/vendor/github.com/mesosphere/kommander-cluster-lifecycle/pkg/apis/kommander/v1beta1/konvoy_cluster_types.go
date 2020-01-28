/*
 * Copyright 2019 Mesosphere, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package v1beta1

import (
	"github.com/mesosphere/konvoy/pkg/apis/konvoy/v1beta1"
	"github.com/mesosphere/konvoy/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	SSHAuthPublicKey = "ssh-publickey"

	KindKonvoyCluster        = "KonvoyCluster"
	KindCloudProviderAccount = "CloudProviderAccount"

	// SecretTypeProviderAWS contains data needed to authenticate to AWS.
	//
	// Required field:
	// - Secret.Data["credentials"] - credentials needed for authentication
	// Optional field:
	// - Secret.Data["config"] - config needed for authentication
	// - Secret.Data["profile"] - profile in credentials and config to use
	SecretTypeProviderAWS corev1.SecretType = "kommander.mesosphere.io/aws-credentials"

	// SecretKeyProviderAWSCredentials is the key for credentials in a SecretTypeProviderAWS
	// type secret.
	SecretKeyProviderAWSCredentials = "credentials"
	// SecretKeyProviderAWSConfig is the key for config in a SecretTypeProviderAWS
	// type secret.
	SecretKeyProviderAWSConfig = "config"
	// SecretKeyProviderAWSProfile is the key for profile in a SecretTypeProviderAWS
	// type secret.
	SecretKeyProviderAWSProfile = "profile"

	// SecretTypeProviderAzure contains data needed to authenticate to Azure.
	//
	// Required fields:
	// - Secret.Data["clientID"] - client ID
	// - Secret.Data["clientSecret"] - client secret
	// - Secret.Data["subscriptionID"] - subscription ID
	// - Secret.Data["tenantID"] - account tenant ID
	SecretTypeProviderAzure corev1.SecretType = "kommander.mesosphere.io/azure-credentials"

	// SecretKeyProviderAzureClientID is the key for client ID in a SecretTypeProviderAzure
	// type secret.
	SecretKeyProviderAzureClientID = "clientID"
	// SecretKeyProviderAzureClientSecret is the key for client secret in a SecretTypeProviderAzure
	// type secret.
	SecretKeyProviderAzureClientSecret = "clientSecret"
	// SecretKeyProviderAzureSubscriptionID is the key for subscription ID in a SecretTypeProviderAzure
	// type secret.
	SecretKeyProviderAzureSubscriptionID = "subscriptionID"
	// SecretKeyProviderAzureTenantID is the key for tenant ID in a SecretTypeProviderAzure
	// type secret.
	SecretKeyProviderAzureTenantID = "tenantID"

	// SecretTypeProviderNone contains data needed to authenticate to Konvoy clusters provisioned with the
	// None provider, i.e. on-prem deployments.
	//
	// Required field:
	// - Secret.Data["ssh-privatekey"] - credentials needed for authentication
	SecretTypeProviderNone corev1.SecretType = "kommander.mesosphere.io/none-credentials"

	// SecretKeyProviderNoneSSHAuthPrivateKey is the key for SSH private key in a SecretTypeProviderNone
	// type secret.
	SecretKeyProviderNoneSSHAuthPrivateKey = corev1.SSHAuthPrivateKey

	// SecretTypeKubeconfig contains data needed to authenticate to a Kubernetes cluster.
	//
	// Required field:
	// - Secret.Data["kubeconfig"] - kubeconfig file for Kubernetes clusters
	SecretTypeKubeconfig corev1.SecretType = "kommander.mesosphere.io/kubeconfig"
	// SecretKeyKubeconfig is the key for admin.conf in a SecretTypeKubeconfig
	// type secret.
	SecretKeyKubeconfig = "kubeconfig"

	// ConfigMapKeyInventoryYAML is the key in a configmap containing the required inventory.yaml file contents.
	ConfigMapKeyInventoryYAML = constants.DefaultInventoryFileName
)

type KonvoyClusterUpgradeStrategy string

const (
	// UpgradeStrategySafe uses a default upgrade strategy for Konvoy
	UpgradeStrategySafe KonvoyClusterUpgradeStrategy = "Safe"
	// UpgradeStrategyForce forces the upgrade of all the nodes of a Konvoy cluster
	UpgradeStrategyForce KonvoyClusterUpgradeStrategy = "Force"
)

// KonvoyClusterSpec defines the desired state of KonvoyCluster
type KonvoyClusterSpec struct {
	ClusterConfiguration     v1beta1.ClusterConfigurationSpec `json:"cluster"`
	ProvisionerConfiguration v1beta1.ClusterProvisionerSpec   `json:"provisioner"`
	CloudProviderAccountRef  corev1.LocalObjectReference      `json:"cloudProviderAccountRef"`
	InventoryRef             *corev1.LocalObjectReference     `json:"inventoryRef,omitempty"`
	UpgradeStrategy          *KonvoyClusterUpgradeStrategy    `json:"upgradeStrategy,omitempty"`
}

// ProviderType is an aliased type for provider types.
type ProviderType string

const (
	// ProviderTypeAWS is the AWS provider type.
	ProviderTypeAWS ProviderType = constants.ProvisionerAWS
	// ProviderTypeAzure is the Azure provider type.
	ProviderTypeAzure ProviderType = constants.ProvisionerAzure
	// ProviderTypeNone is the None provider type (used for on-prem).
	ProviderTypeNone ProviderType = constants.ProvisionerNone
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=cpa;cpas
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display name",type="string",JSONPath=".metadata.annotations.kommander\\.mesosphere\\.io/display-name"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CloudProviderAccount holds the details for specific cloud provider accounts.
type CloudProviderAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudProviderAccountSpec   `json:"spec,omitempty"`
	Status CloudProviderAccountStatus `json:"status,omitempty"`
}

// CloudProviderAccountSpec defines the state for the CloudProviderAccount.
type CloudProviderAccountSpec struct {
	Provider       ProviderType                `json:"provider"`
	CredentialsRef corev1.LocalObjectReference `json:"credentialsRef"`
}

// CloudProviderAccountSpec holds the status for the CloudProviderAccount.
// Currently empty.
// TODO: Define phase and/or conditions that are useful to clients, such as invalid credentials, expired credentials, etc
type CloudProviderAccountStatus struct {
}

// +kubebuilder:object:root=true

// CloudProviderAccountList contains a list of CloudProviderAccount
type CloudProviderAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudProviderAccount `json:"items"`
}

// KonvoyClusterPhase is a string representation of a KonvoyCluster Phase.
//
// This type is a high-level indicator of the status of the KonvoyCluster as it is provisioned,
// from the API user’s perspective.
//
// The value should not be interpreted by any software components as a reliable indication
// of the actual state of the KonvoyCluster, and controllers should not use the KonvoyCluster Phase field
// value when making decisions about what action to take.
//
// Controllers should always look at the actual state of the KonvoyCluster’s fields to make those decisions.
type KonvoyClusterPhase string

var (
	// KonvoyClusterPhasePending is the first state a KonvoyCluster is assigned by
	// Cluster API KonoyCluster controller after being created.
	KonvoyClusterPhasePending = KonvoyClusterPhase("Pending")

	// KonvoyClusterPhaseProvisioning is the state when the KonvoyCluster has a provider infrastructure
	// object associated and can start provisioning.
	KonvoyClusterPhaseProvisioning = KonvoyClusterPhase("Provisioning")

	// KonvoyClusterPhaseProvisioned is the state when its
	// infrastructure has been created and configured.
	KonvoyClusterPhaseProvisioned = KonvoyClusterPhase("Provisioned")

	// KonvoyClusterPhaseDeleting is the KonvoyCluster state when a delete
	// request has been sent to the API Server,
	// but its infrastructure has not yet been fully deleted.
	KonvoyClusterPhaseDeleting = KonvoyClusterPhase("Deleting")

	// KonvoyClusterPhaseDeleteFailed is the KonvoyCluster state when the system
	// might require user intervention.
	KonvoyClusterPhaseDeleteFailed = KonvoyClusterPhase("DeleteFailed")

	// KonvoyClusterPhaseFailed is the KonvoyCluster state when the system
	// might require user intervention.
	KonvoyClusterPhaseFailed = KonvoyClusterPhase("Failed")

	// KonvoyClusterPhaseDeleted is the KonvoyCluster state when the system
	// might require user intervention.
	KonvoyClusterPhaseDeleted = KonvoyClusterPhase("Deleted")

	// KonvoyClusterPhaseUnknown is returned if the KonvoyCluster state cannot be determined.
	KonvoyClusterPhaseUnknown = KonvoyClusterPhase("")
)

// KonvoyClusterStatus defines the observed state of KonvoyCluster
type KonvoyClusterStatus struct {
	// Phase represents the current phase of Konvoy cluster actuation.
	// E.g. Pending, Provisioning, Provisioned, Deleting, Failed, etc.
	// +optional
	Phase KonvoyClusterPhase `json:"phase,omitempty"`

	// AdminConfSecretRef holds a reference to the admin conf secret when it is available.
	AdminConfSecretRef *corev1.LocalObjectReference `json:"adminConfSecretRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display name",type="string",JSONPath=".metadata.annotations.kommander\\.mesosphere\\.io/display-name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provisioner.provider"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=konvoy;konvoys

// KonvoyCluster is the Schema for the Konvoy cluster kind
type KonvoyCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KonvoyClusterSpec   `json:"spec,omitempty"`
	Status KonvoyClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KonvoyClusterList contains a list of KonvoyCluster
type KonvoyClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KonvoyCluster `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addKnownKonvoyClusterTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownKonvoyClusterTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&KonvoyCluster{},
		&KonvoyClusterList{},
		&CloudProviderAccount{},
		&CloudProviderAccountList{},
	)
	return nil
}
