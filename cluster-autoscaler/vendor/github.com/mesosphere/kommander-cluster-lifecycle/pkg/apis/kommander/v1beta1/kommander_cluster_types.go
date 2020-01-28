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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	KindKommanderCluster = "KommanderCluster"
)

// KommanderClusterSpec defines the desired state of Cluster
type KommanderClusterSpec struct {
	KubeconfigRef *corev1.LocalObjectReference `json:"kubeconfigRef,omitempty"`
	ClusterRef    *ClusterReference            `json:"clusterRef,omitempty"`
}

// ClusterReference holds a single reference to clusters provisioned via Kommander.
// Only one field is allowed to be set. Currently, only Konvoy clusters are creatable,
// but this is left extensible for other provider types in the future.
type ClusterReference struct {
	KonvoyCluster *corev1.LocalObjectReference `json:"konvoyCluster,omitempty"`
}

type KommanderClusterPhase string

var (
	// KommanderClusterPhasePending is the first state a KommanderCluster is assigned by
	// Cluster API KommanderCluster controller after being created.
	KommanderClusterPhasePending = KommanderClusterPhase("Pending")

	// KommanderClusterPhaseJoining is the state when the KommanderCluster has a valid kubeconfig
	// associated and can start joining.
	KommanderClusterPhaseJoining = KommanderClusterPhase("Joining")

	// KommanderClusterPhaseJoined is the KommanderCluster state when it has been joined
	// to kubefed successfully.
	KommanderClusterPhaseJoined = KommanderClusterPhase("Joined")

	// KommanderClusterPhaseJoinFailed is the KommanderCluster state when joining failed.
	KommanderClusterPhaseJoinFailed = KommanderClusterPhase("JoinFailed")

	// KommanderClusterPhaseUnjoining is the state when the KommanderCluster is being deleted
	// and is being unjoined from kubefed.
	KommanderClusterPhaseUnjoining = KommanderClusterPhase("Unjoining")

	// KommanderClusterPhaseUnjoined is the KommanderCluster state when a unjoining
	// has successfully completed.
	KommanderClusterPhaseUnjoined = KommanderClusterPhase("Unjoined")

	// KommanderClusterPhaseUnjoinFailed is the KommanderCluster state when a unjoining
	// has failed.
	KommanderClusterPhaseUnjoinFailed = KommanderClusterPhase("UnjoinFailed")

	// KommanderClusterPhaseUnknown is returned if the KommanderCluster state cannot be determined.
	KommanderClusterPhaseUnknown = KommanderClusterPhase("")
)

// KommanderClusterStatus defines the observed state of Cluster
type KommanderClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase represents the current phase of cluster actuation.
	// E.g. Pending, Provisioning, Provisioned, Deleting, Failed, etc.
	// +optional
	Phase KommanderClusterPhase `json:"phase,omitempty"`

	// KubefedClusterRef holds a reference to a kubefedcluster in the
	// kubefed system namespace.
	// +optional
	KubefedClusterRef *corev1.LocalObjectReference `json:"kubefedclusterRef,omitempty"`

	// DexTFAClientRef holds a reference to a DexClient provisioned for
	// Traefik Forward Auth running on managed cluster.
	DexTFAClientRef *corev1.ObjectReference `json:"dextfaclientRef,omitempty"`

	// ServiceEndpoints will be the addresses assigned to the Kubernetes exposed services
	// +optional
	ServiceEndpoints map[string]string `json:"serviceEndpoints,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display name",type="string",JSONPath=".metadata.annotations.kommander\\.mesosphere\\.io/display-name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Kubefed Cluster",type="string",JSONPath=".status.kubefedclusterRef.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=komm;komms;kommander;kommanders

// KommanderCluster is the Schema for the kommander clusters API
type KommanderCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KommanderClusterSpec   `json:"spec,omitempty"`
	Status KommanderClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type KommanderClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KommanderCluster `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addKnownKommanderClusterTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownKommanderClusterTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &KommanderCluster{}, &KommanderClusterList{})
	return nil
}
