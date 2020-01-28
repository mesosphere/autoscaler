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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	KindKonvoyManagementCluster = "KonvoyMangementCluster"
)

// KonvoyManagementClusterSpec defines the desired state of KonvoyManagementCluster
type KonvoyManagementClusterSpec struct {
	ClusterConfiguration     v1beta1.ClusterConfigurationSpec `json:"cluster"`
	ProvisionerConfiguration v1beta1.ClusterProvisionerSpec   `json:"provisioner"`
	CloudProviderAccountRef  corev1.LocalObjectReference      `json:"cloudProviderAccountRef"`
	InventoryRef             *corev1.LocalObjectReference     `json:"inventoryRef,omitempty"`
	UpgradeStrategy          *KonvoyClusterUpgradeStrategy    `json:"upgradeStrategy,omitempty"`
	AdminConfSecretRef       *corev1.LocalObjectReference     `json:"adminConfSecretRef,omitempty"`
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

// KonvoyManagementClusterStatus defines the observed state of KonvoyManagementCluster
type KonvoyManagementClusterStatus struct {
	// Phase represents the current phase of Konvoy cluster actuation.
	// E.g. Pending, Provisioning, Provisioned, Deleting, Failed, etc.
	// +optional
	Phase KonvoyClusterPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display name",type="string",JSONPath=".metadata.annotations.kommander\\.mesosphere\\.io/display-name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provisioner.provider"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=konvoymanagementcluster;konvoymanagementclusters

// KonvoyManagementCluster is the Schema for the Konvoy cluster kind
type KonvoyManagementCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KonvoyManagementClusterSpec   `json:"spec,omitempty"`
	Status KonvoyManagementClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KonvoyManagementClusterList contains a list of KonvoyManagementCluster
type KonvoyManagementClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KonvoyManagementCluster `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addKnownKonvoyManagementClusterTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownKonvoyManagementClusterTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&KonvoyManagementCluster{},
		&KonvoyManagementClusterList{},
		&CloudProviderAccount{},
		&CloudProviderAccountList{},
	)
	return nil
}
