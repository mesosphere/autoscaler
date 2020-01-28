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
	"k8s.io/utils/pointer"

	"github.com/mesosphere/konvoy/pkg/constants"
)

func SetDefaults_ClusterConfigurationSpec(obj *ClusterConfigurationSpec) { //nolint:stylecheck
	if obj.Kubernetes == nil {
		obj.Kubernetes = &Kubernetes{}
	}
	if obj.ContainerNetworking == nil {
		obj.ContainerNetworking = &ContainerNetworking{}
	}
	if obj.ContainerRuntime == nil {
		obj.ContainerRuntime = &ContainerRuntime{}
	}
	if obj.NodePools == nil {
		obj.NodePools = []NodePool{{
			Name: "worker",
		}}
	}
	if obj.OSPackages == nil {
		obj.OSPackages = &OSPackages{
			EnableAdditionalRepositories: pointer.BoolPtr(constants.DefaultEnableAdditionalRepositories),
		}
	}
	if obj.Version == nil && constants.KonvoyVersion != "" {
		obj.Version = pointer.StringPtr(constants.KonvoyVersion)
	}
}

func SetDefaults_Kubernetes(obj *Kubernetes) { //nolint:stylecheck
	if obj.Version == nil && constants.DefaultKubernetesVersion != "" {
		obj.Version = pointer.StringPtr(constants.DefaultKubernetesVersion)
	}
	if obj.Networking == nil {
		obj.Networking = &Networking{}
	}
	if obj.CloudProvider == nil {
		obj.CloudProvider = &CloudProvider{}
	}
	if obj.AdmissionPlugins == nil {
		obj.AdmissionPlugins = &AdmissionPlugins{}
	}
}

func SetDefaults_Networking(obj *Networking) { //nolint:stylecheck
	if obj.PodSubnet == nil {
		obj.PodSubnet = pointer.StringPtr(constants.DefaultPodSubnet)
	}
	if obj.ServiceSubnet == nil {
		obj.ServiceSubnet = pointer.StringPtr(constants.DefaultServiceSubnet)
	}
}

func SetDefaults_CloudProvider(obj *CloudProvider) { //nolint:stylecheck
	if obj.Provider == nil {
		obj.Provider = pointer.StringPtr(constants.ProvisionerNone)
	}
}

func SetDefaults_AdmissionPlugins(obj *AdmissionPlugins) { //nolint:stylecheck
	if obj.Enabled == nil {
		obj.Enabled = constants.DefaultEnabledAdmissionPlugins
	}
}

func SetDefaults_ContainerNetworking(obj *ContainerNetworking) { //nolint:stylecheck
	if obj.Calico == nil {
		obj.Calico = &CalicoContainerNetworking{}
	}
}

func SetDefaults_CalicoContainerNetworking(obj *CalicoContainerNetworking) { //nolint:stylecheck
	if obj.Version == nil {
		obj.Version = pointer.StringPtr(constants.DefaultCalicoVersion)
	}
}

func SetDefaults_ContainerRuntime(obj *ContainerRuntime) { //nolint:stylecheck
	if obj.Containerd == nil {
		obj.Containerd = &ContainerdContainerRuntime{}
	}
}

func SetDefaults_ContainerdContainerRuntime(obj *ContainerdContainerRuntime) { //nolint:stylecheck
	if obj.Version == nil {
		obj.Version = pointer.StringPtr(constants.DefaultContainerdVersion)
	}
}
