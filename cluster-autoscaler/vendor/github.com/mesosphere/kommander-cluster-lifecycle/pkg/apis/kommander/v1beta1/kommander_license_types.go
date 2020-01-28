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
	KindKommanderLicense = "License"
)

// LicenseSpec defines the desired state of License
type LicenseSpec struct {
	// LicenseReference holds a single reference to the secret holding the license JWT
	LicenseRef *corev1.LocalObjectReference `json:"licenseRef,omitempty"`
}

// LicenseStatus defines the observed state of License
type LicenseStatus struct {
	// Indicates whether the license is valid, i.e. the secret containing the JWT exists and the JWT carries a valid
	// D2iQ signature. This does NOT indicate whether the license has expired or terms have been breached.
	Valid bool `json:"valid"`
	// The customer's ID. This is the customer name provided from Salesforce.
	CustomerID string `json:"customerId"`
	// The license's ID as provided from Salesforce.
	LicenseID string `json:"licenseId"`
	// Start date of the licensing period.
	StartDate *metav1.Time `json:"startDate,omitempty"`
	// End date of the licensing period.
	EndDate *metav1.Time `json:"endDate,omitempty"`
	// Maximum number of clusters that the license allows.
	ClusterCapacity int `json:"clusterCapacity"`
	// The license's version ID as provided when the license was created.
	Version string `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Valid",type="string",JSONPath=".status.valid"
// +kubebuilder:printcolumn:name="Start date",type="string",JSONPath=".status.startDate"
// +kubebuilder:printcolumn:name="End date",type="string",JSONPath=".status.endDate"
// +kubebuilder:printcolumn:name="Cluster capacity",type="string",JSONPath=".status.clusterCapacity"
// +kubebuilder:printcolumn:name="JWT reference",type="string",JSONPath=".spec.licenseRef.name"

// License is the Schema for the licenses API
type License struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LicenseSpec   `json:"spec,omitempty"`
	Status LicenseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LicenseList contains a list of License
type LicenseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []License `json:"items"`
}

func init() {
	localSchemeBuilder.Register(addKnownLicenseTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownLicenseTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &License{}, &LicenseList{})
	return nil
}
