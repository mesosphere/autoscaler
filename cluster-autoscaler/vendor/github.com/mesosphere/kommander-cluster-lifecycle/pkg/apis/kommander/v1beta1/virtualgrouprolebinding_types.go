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

// VirtualGroupRoleBindingSpec defines the desired state of VirtualGroupRoleBinding
type VirtualGroupRoleBindingSpec struct {
	// TODO coordinate with frontend team to change this to a rbacv1.RoleRef
	// type
	RoleRef         *corev1.LocalObjectReference `json:"roleRef,omitempty"`
	VirtualGroupRef *corev1.LocalObjectReference `json:"virtualGroupRef,omitempty"`
	Placement       PlacementSelector            `json:"placement,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// VirtualGroupRoleBinding is the Schema for the virtualgrouprolebindings API
type VirtualGroupRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualGroupRoleBindingSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualGroupRoleBindingList contains a list of VirtualGroupRoleBinding
type VirtualGroupRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualGroupRoleBinding `json:"items"`
}

// GetClusterNames returns the list of cluster names
func (v *VirtualGroupRoleBinding) GetClusterNames() []string {
	clusterNames := make([]string, 0, len(v.Spec.Placement.Clusters))
	for _, cluster := range v.Spec.Placement.Clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}

	return clusterNames
}

func init() {
	localSchemeBuilder.Register(addKnownVirtualGroupRoleBindingTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownVirtualGroupRoleBindingTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &VirtualGroupRoleBinding{}, &VirtualGroupRoleBindingList{})
	return nil
}
