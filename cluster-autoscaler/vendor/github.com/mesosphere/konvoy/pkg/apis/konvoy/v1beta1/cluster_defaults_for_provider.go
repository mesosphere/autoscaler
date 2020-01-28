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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/pointer"

	"github.com/mesosphere/konvoy/pkg/constants"
)

func SetObjectDefaultsClusterConfigurationForProvider(provider string, in *ClusterConfiguration) {
	SetDefaultsClusterConfigurationSpecForProvider(provider, &in.Spec)
	SetObjectDefaults_ClusterConfiguration(in)
}

func SetDefaultsClusterConfigurationSpecForProvider(provider string, obj *ClusterConfigurationSpec) {
	if obj.ContainerNetworking == nil {
		obj.ContainerNetworking = &ContainerNetworking{}
	}
	SetDefaultsContainerNetworkingForProvider(provider, obj.ContainerNetworking)
	if obj.Kubernetes == nil {
		obj.Kubernetes = &Kubernetes{}
	}
	SetDefaultsKubernetesForProvider(provider, obj.Kubernetes)
}

func SetDefaultsContainerNetworkingForProvider(provider string, obj *ContainerNetworking) {
	if obj.Calico == nil {
		obj.Calico = &CalicoContainerNetworking{}
	}
	SetDefaultsCalicoContainerNetworkingForProvider(provider, obj.Calico)
}

func SetDefaultsCalicoContainerNetworkingForProvider(provider string, obj *CalicoContainerNetworking) {
	if obj.Encapsulation == nil {
		switch provider {
		case constants.ProvisionerAWS:
			obj.Encapsulation = pointer.StringPtr(constants.CalicoEncapsulationModeIPIP)
		case constants.ProvisionerAzure:
			obj.Encapsulation = pointer.StringPtr("")
		case constants.ProvisionerNone:
			obj.Encapsulation = pointer.StringPtr(constants.CalicoEncapsulationModeIPIP)
		case constants.ProvisionerDocker:
			obj.Encapsulation = pointer.StringPtr(constants.CalicoEncapsulationModeIPIP)
		default:
			// Noop.
		}
	}
	if obj.MTU == nil {
		// The following defaults for MTU are derived from:
		// https://docs.projectcalico.org/master/networking/mtu
		var encapOverhead int32
		if obj.Encapsulation != nil {
			switch *obj.Encapsulation {
			case constants.CalicoEncapsulationModeIPIP:
				encapOverhead = constants.CalicoEncapsulationModeIPIPOverhead
			case constants.CalicoEncapsulationModeVXLAN:
				encapOverhead = constants.CalicoEncapsulationModeVXLANOverhead
			}
		}

		switch provider {
		case constants.ProvisionerAWS:
			obj.MTU = pointer.Int32Ptr(constants.DefaultMTUProvisionerAWS - encapOverhead)
		case constants.ProvisionerAzure:
			obj.MTU = pointer.Int32Ptr(constants.DefaultMTUProvisionerAzure - encapOverhead)
		case constants.ProvisionerNone:
			obj.MTU = pointer.Int32Ptr(constants.DefaultMTUProvisionerNone - encapOverhead)
		case constants.ProvisionerDocker:
			obj.MTU = pointer.Int32Ptr(constants.DefaultMTUProvisionerDocker - encapOverhead)
		default:
			// Noop.
		}
	}
}

func SetDefaultsKubernetesForProvider(provider string, obj *Kubernetes) {
	switch provider {
	case constants.ProvisionerAWS:
		if obj.CloudProvider == nil {
			obj.CloudProvider = &CloudProvider{}
		}
		SetDefaultsCloudProviderForProvider(provider, obj.CloudProvider)
	case constants.ProvisionerAzure:
		if obj.CloudProvider == nil {
			obj.CloudProvider = &CloudProvider{}
		}
		SetDefaultsCloudProviderForProvider(provider, obj.CloudProvider)
		if obj.Networking == nil {
			obj.Networking = &Networking{}
		}
		SetDefaultsNetworkingForProvider(provider, obj.Networking)
	case constants.ProvisionerDocker:
		if obj.Networking == nil {
			obj.Networking = &Networking{}
		}
		SetDefaultsNetworkingForProvider(provider, obj.Networking)
		if obj.PreflightChecks == nil {
			obj.PreflightChecks = &PreflightChecks{}
		}
		SetDefaultsPreflightChecksForProvider(provider, obj.PreflightChecks)
		if obj.Kubelet == nil {
			obj.Kubelet = &Kubelet{}
		}
		SetDefaultsKubeletForProvider(provider, obj.Kubelet)
	default:
		// Noop.
	}
}

func SetDefaultsCloudProviderForProvider(provider string, obj *CloudProvider) {
	switch provider {
	case constants.ProvisionerAWS:
		if obj.Provider == nil {
			obj.Provider = pointer.StringPtr(constants.ProvisionerAWS)
		}
	case constants.ProvisionerAzure:
		if obj.Provider == nil {
			obj.Provider = pointer.StringPtr(constants.ProvisionerAzure)
		}
	default:
		// Noop.
	}
}

func SetDefaultsControlPlaneForProvider(provider string, obj *ControlPlane) {
	switch provider {
	case constants.ProvisionerAWS:
		// Noop.
	case constants.ProvisionerDocker:
		if obj.Keepalived == nil {
			obj.Keepalived = &Keepalived{}
		}
		if obj.ControlPlaneEndpointOverride == nil {
			obj.ControlPlaneEndpointOverride = pointer.StringPtr(constants.DefaultDockerControlPlaneEndpointOverride)
		}
	case constants.ProvisionerNone:
		if obj.Keepalived == nil {
			obj.Keepalived = &Keepalived{}
		}
		if obj.ControlPlaneEndpointOverride == nil {
			obj.ControlPlaneEndpointOverride = pointer.StringPtr(constants.DefaultControlPlaneEndpointOverride)
		}
		if obj.Certificate == nil {
			obj.Certificate = &Certificate{}
		}
	default:
		// Noop.
	}
}

func SetDefaultsNetworkingForProvider(provider string, obj *Networking) {
	switch provider {
	case constants.ProvisionerAzure:
		if obj.PodSubnet == nil {
			obj.PodSubnet = pointer.StringPtr(constants.DefaultAzurePodSubnet)
		}
	case constants.ProvisionerDocker:
		if obj.PodSubnet == nil {
			obj.PodSubnet = pointer.StringPtr(constants.DefaultDockerPodSubnet)
		}
		if obj.ServiceSubnet == nil {
			obj.ServiceSubnet = pointer.StringPtr(constants.DefaultDockerServiceSubnet)
		}
	default:
		// Noop.
	}
}

func SetDefaultsPreflightChecksForProvider(provider string, obj *PreflightChecks) {
	switch provider {
	case constants.ProvisionerDocker:
		// We have to ignore some preflight checks when running
		// kubernetes in Docker so that we don't get errors on OSX and
		// most of the Linux distros. For `SystemVerification` check,
		// it will throw error on some old Linux distro that does not
		// serve `/proc/config.gz`.
		existingPreflightChecksToIgnore := sets.NewString(obj.ErrorsToIgnore...)
		existingPreflightChecksToIgnore.Insert(
			"swap",
			"SystemVerification",
			"filecontent--proc-sys-net-bridge-bridge-nf-call-iptables",
			"filecontent--proc-sys-net-bridge-bridge-nf-call-ip6tables",
		)
		obj.ErrorsToIgnore = existingPreflightChecksToIgnore.List()
	default:
		// Noop.
	}
}

func SetDefaultsKubeletForProvider(provider string, obj *Kubelet) {
	switch provider {
	case constants.ProvisionerDocker:
		obj.CgroupRoot = pointer.StringPtr(constants.DefaultDockerKubeletCgroupRoot)
	default:
		// Noop.
	}
}
