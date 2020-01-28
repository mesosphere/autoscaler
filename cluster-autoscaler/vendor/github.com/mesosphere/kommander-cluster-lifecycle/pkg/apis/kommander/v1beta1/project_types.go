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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kubefed/pkg/controller/util"
)

const (
	KindProject = "Project"
)

// +genclient:nonNamespaced

// +kubebuilder:object:root=true

// ProjectList is a list of Project objects.
type ProjectList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []Project
}

const (
	// These are internal finalizer values to Project API.
	FinalizerProject core.FinalizerName = "kommander.mesosphere.com/project"
)

// ProjectSpec describes the attributes on a Project
type ProjectSpec struct {
	Placement PlacementSelector `json:"placement,omitempty"`
}

// PlacementSelector defines the fields to select clusters where to apply the project
type PlacementSelector struct {
	Clusters        []GenericClusterReference `json:"clusters,omitempty"`
	ClusterSelector *metav1.LabelSelector     `json:"clusterSelector,omitempty"`
}

// GenericClusterReference sets the name of the cluster
type GenericClusterReference struct {
	Name string `json:"name"`
}

// ProjectStatus is information about the current status of a Project
type ProjectStatus struct {
	Phase core.NamespacePhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:path=projects,scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display name",type="string",JSONPath=".metadata.annotations.kommander\\.mesosphere\\.io/display-name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"

// Project is a logical top-level container for a set of Kommander resources
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

// GetClusterNames returns the list of cluster names
func (project *Project) GetClusterNames() []string {
	clusterNames := make([]string, 0, len(project.Spec.Placement.Clusters))
	for _, cluster := range project.Spec.Placement.Clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}

	return clusterNames
}

// GetClusterSelector returns the label selector for all the cluster labels
func (project *Project) GetClusterSelector() (labels.Selector, error) {
	return metav1.LabelSelectorAsSelector(project.Spec.Placement.ClusterSelector)
}

func ConvertPlacementToKubefedPlacement(p PlacementSelector) util.GenericPlacementFields {
	clusters := make([]util.GenericClusterReference, 0, len(p.Clusters))
	for _, c := range p.Clusters {
		clusters = append(clusters, util.GenericClusterReference{Name: c.Name})
	}
	return util.GenericPlacementFields{
		Clusters:        clusters,
		ClusterSelector: p.ClusterSelector,
	}
}

func init() {
	localSchemeBuilder.Register(addKnownProjectTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownProjectTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&Project{},
		&ProjectList{},
	)
	return nil
}
