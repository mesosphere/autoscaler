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

// VirtualGroupClusterRoleBindingSpec defines the desired state of VirtualGroupClusterRoleBinding
type VirtualGroupClusterRoleBindingSpec struct {
	// TODO coordinate with frontend team to change this to a rbacv1.RoleRef
	// type
	ClusterRoleRef  *corev1.LocalObjectReference `json:"clusterRoleRef,omitempty"`
	VirtualGroupRef *corev1.LocalObjectReference `json:"virtualGroupRef,omitempty"`
	Placement       PlacementSelector            `json:"placement,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=virtualgroupclusterrolebindings,scope=Cluster
// +kubebuilder:storageversion

// VirtualGroupClusterRoleBinding is the Schema for the virtualgroupclusterrolebindings API
type VirtualGroupClusterRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualGroupClusterRoleBindingSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualGroupClusterRoleBindingList contains a list of VirtualGroupClusterRoleBinding
type VirtualGroupClusterRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualGroupClusterRoleBinding `json:"items"`
}

// GetClusterNames returns the list of cluster names
func (v *VirtualGroupClusterRoleBinding) GetClusterNames() []string {
	clusterNames := make([]string, 0, len(v.Spec.Placement.Clusters))
	for _, cluster := range v.Spec.Placement.Clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}

	return clusterNames
}

func init() {
	localSchemeBuilder.Register(addKnownVirtualGroupClusterRoleBindingTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownVirtualGroupClusterRoleBindingTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &VirtualGroupClusterRoleBinding{}, &VirtualGroupClusterRoleBindingList{})
	return nil
}
