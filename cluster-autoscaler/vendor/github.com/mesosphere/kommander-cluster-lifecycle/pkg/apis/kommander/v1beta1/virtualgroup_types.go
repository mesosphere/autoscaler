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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	KindVirtualGroup = "VirtualGroup"
)

// VirtualGroupSpec defines the desired state of VirtualGroup
type VirtualGroupSpec struct {
	Subjects []rbacv1.Subject `json:"subjects,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=virtualgroups,scope=Cluster
// +kubebuilder:storageversion

// VirtualGroup is the Schema for the virtualgroups API
type VirtualGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualGroupSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualGroupList contains a list of VirtualGroup
type VirtualGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualGroup `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addKnownVirtualGroupTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownVirtualGroupTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &VirtualGroup{}, &VirtualGroupList{})
	return nil
}
