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

// +k8s:defaulter-gen=covers
func SetDefaults_ClusterProvisionerSpec(obj *ClusterProvisionerSpec) { //nolint:stylecheck
	if obj.Version == nil {
		obj.Version = pointer.StringPtr(constants.KonvoyVersion)
	}

	switch obj.Provider {
	case constants.ProvisionerAWS:
		if obj.AWSProviderOptions == nil {
			obj.AWSProviderOptions = &AWSProviderOptions{}
		}
		SetDefaults_AWSProviderOptions(obj.AWSProviderOptions)
		if obj.NodePools == nil {
			obj.NodePools = []MachinePool{{
				Name:    "worker",
				Count:   constants.DefaultAWSWorkerCount,
				Machine: &Machine{},
			}, {
				Name:         "control-plane",
				ControlPlane: pointer.BoolPtr(true),
				Count:        constants.DefaultAWSControlPlaneCount,
				Machine: &Machine{
					Type:           pointer.StringPtr(constants.DefaultAWSControlPlaneMachineType),
					RootVolumeType: pointer.StringPtr(constants.DefaultAWSControlPlaneMachineRootVolumeType),
					RootVolumeIOPS: pointer.Int32Ptr(constants.DefaultAWSControlPlaneMachineRootVolumeIOPS),
				},
			}, {
				Name:    "bastion",
				Bastion: pointer.BoolPtr(true),
				Count:   0,
				Machine: &Machine{
					Type:                 pointer.StringPtr(constants.DefaultAWSBastionMachineType),
					RootVolumeSize:       pointer.Int64Ptr(constants.DefaultAWSBastionMachineRootVolumeSize),
					ImagefsVolumeEnabled: pointer.BoolPtr(false),
				},
			}}
		}
		for i := range obj.NodePools {
			nodePool := &obj.NodePools[i]
			SetDefaults_MachinePool_AWS(nodePool)
			if nodePool.Machine != nil {
				SetDefaults_Machine_AWS(nodePool.Machine)
			}
		}
	case constants.ProvisionerAzure:
		if obj.AzureProviderOptions == nil {
			obj.AzureProviderOptions = &AzureProviderOptions{}
		}
		SetDefaults_AzureProviderOptions(obj.AzureProviderOptions)
		if obj.NodePools == nil {
			obj.NodePools = []MachinePool{{
				Name:    "worker",
				Count:   constants.DefaultAzureWorkerCount,
				Machine: &Machine{},
			}, {
				Name:         "control-plane",
				ControlPlane: pointer.BoolPtr(true),
				Count:        constants.DefaultAzureControlPlaneCount,
				Machine: &Machine{
					Type: pointer.StringPtr(constants.DefaultAzureControlPlaneMachineType),
				},
			}, {
				Name:    "bastion",
				Bastion: pointer.BoolPtr(true),
				Count:   0,
				Machine: &Machine{
					Type:                 pointer.StringPtr(constants.DefaultAzureBastionMachineType),
					RootVolumeSize:       pointer.Int64Ptr(constants.DefaultAzureBastionMachineRootVolumeSize),
					ImagefsVolumeEnabled: pointer.BoolPtr(false),
				},
			}}
		}
		for i := range obj.NodePools {
			nodePool := &obj.NodePools[i]
			SetDefaults_MachinePool_Azure(nodePool)
			if nodePool.Machine != nil {
				SetDefaults_Machine_Azure(nodePool.Machine)
			}
		}
	case constants.ProvisionerDocker:
		if obj.DockerProviderOptions == nil {
			obj.DockerProviderOptions = &DockerProviderOptions{}
		}
		SetDefaults_DockerProviderOptions(obj.DockerProviderOptions)
		if obj.NodePools == nil {
			obj.NodePools = []MachinePool{{
				Name:    "worker",
				Count:   constants.DefaultDockerWorkerCount,
				Machine: &Machine{},
			}, {
				Name:         "control-plane",
				ControlPlane: pointer.BoolPtr(true),
				Count:        constants.DefaultDockerControlPlaneCount,
				Machine:      &Machine{},
			}}
		}
		for i := range obj.NodePools {
			nodePool := &obj.NodePools[i]
			SetDefaults_MachinePool_Docker(nodePool)
			if nodePool.Machine != nil {
				SetDefaults_Machine_Docker(nodePool.Machine)
			}
		}
	}
}

func SetDefaults_AWSProviderOptions(obj *AWSProviderOptions) { //nolint:stylecheck
	if obj.Region == nil {
		obj.Region = pointer.StringPtr(constants.DefaultAWSRegion)
	}
	if obj.AvailabilityZones == nil && *obj.Region == constants.DefaultAWSRegion {
		obj.AvailabilityZones = constants.DefaultAWSAvailabilityZones
	}
	if obj.VPC == nil {
		obj.VPC = &VPC{}
	}
	if obj.VPC.EnableInternetGateway == nil {
		obj.VPC.EnableInternetGateway = pointer.BoolPtr(true)
	}
	if obj.VPC.EnableVPCEndpoints == nil {
		obj.VPC.EnableVPCEndpoints = pointer.BoolPtr(true)
	}
}

func SetDefaults_AzureProviderOptions(obj *AzureProviderOptions) { //nolint:stylecheck
	if obj.Location == nil {
		obj.Location = pointer.StringPtr(constants.DefaultAzureLocation)
	}
	if obj.AvailabilitySet == nil && *obj.Location == constants.DefaultAzureLocation {
		obj.AvailabilitySet = &AvailabilitySet{}
		obj.AvailabilitySet.FaultDomainCount = pointer.Int32Ptr(constants.DefaultAzureFaultDomainCount)
		obj.AvailabilitySet.UpdateDomainCount = pointer.Int32Ptr(constants.DefaultAzureUpdateDomainCount)
	}
	if obj.Tags == nil {
		obj.Tags = map[string]string{
			"owner": constants.User,
		}
	}
}

func SetDefaults_DockerProviderOptions(obj *DockerProviderOptions) { //nolint:stylecheck
	if obj.DisablePortMapping == nil {
		// We disable port mapping on Linux because it's not
		// needed. Port mapping is only needed on macOS due to
		// the fact that container IP is not directly
		// reachable from the host.
		obj.DisablePortMapping = pointer.BoolPtr(constants.UserOS == "linux")
	}
	if obj.ControlPlaneMappedPortBase == nil {
		obj.ControlPlaneMappedPortBase = pointer.Int32Ptr(constants.DefaultDockerControlPlaneMappedPortBase)
	}
	if obj.SSHControlPlaneMappedPortBase == nil {
		obj.SSHControlPlaneMappedPortBase = pointer.Int32Ptr(constants.DefaultDockerSSHControlPlaneMappedPortBase)
	}
	if obj.SSHWorkerMappedPortBase == nil {
		obj.SSHWorkerMappedPortBase = pointer.Int32Ptr(constants.DefaultDockerSSHWorkerMappedPortBase)
	}
}

func SetDefaults_MachinePool_Docker(obj *MachinePool) { //nolint:stylecheck
	if obj.Machine == nil {
		obj.Machine = &Machine{}
	}
}

func SetDefaults_Machine_Docker(obj *Machine) { //nolint:stylecheck
	if obj.ImageName == nil {
		obj.ImageName = pointer.StringPtr(constants.DefaultDockerBaseImage + ":" + constants.KonvoyVersion)
	}
}

func SetDefaults_MachinePool_AWS(obj *MachinePool) { //nolint:stylecheck
	if obj.Machine == nil {
		obj.Machine = &Machine{}
	}
}

func SetDefaults_MachinePool_Azure(obj *MachinePool) { //nolint:stylecheck
	if obj.Machine == nil {
		obj.Machine = &Machine{}
	}
}

func SetDefaults_Machine_AWS(obj *Machine) { //nolint:stylecheck
	if obj.Type == nil {
		obj.Type = pointer.StringPtr(constants.DefaultAWSWorkerMachineType)
	}
	if obj.RootVolumeType == nil {
		obj.RootVolumeType = pointer.StringPtr(constants.DefaultAWSMachineRootVolumeType)
	}
	if obj.RootVolumeSize == nil {
		obj.RootVolumeSize = pointer.Int64Ptr(constants.DefaultAWSMachineRootVolumeSize)
	}
	if obj.ImagefsVolumeEnabled == nil {
		obj.ImagefsVolumeEnabled = pointer.BoolPtr(true)
	}
	if obj.ImagefsVolumeEnabled != nil && *obj.ImagefsVolumeEnabled {
		if obj.ImagefsVolumeType == nil {
			obj.ImagefsVolumeType = pointer.StringPtr(constants.DefaultAWSMachineImagefsVolumeType)
		}
		if obj.ImagefsVolumeSize == nil {
			obj.ImagefsVolumeSize = pointer.Int64Ptr(constants.DefaultAWSMachineImagefsVolumeSize)
		}
		if obj.ImagefsVolumeDevice == nil {
			obj.ImagefsVolumeDevice = pointer.StringPtr(constants.DefaultAWSMachineImagefsVolumeDevice)
		}
	}
}

func SetDefaults_Machine_Azure(obj *Machine) { //nolint:stylecheck
	if obj.Type == nil {
		obj.Type = pointer.StringPtr(constants.DefaultAzureWorkerMachineType)
	}
	if obj.RootVolumeType == nil {
		obj.RootVolumeType = pointer.StringPtr(constants.DefaultAzureMachineRootVolumeType)
	}
	if obj.RootVolumeSize == nil {
		obj.RootVolumeSize = pointer.Int64Ptr(constants.DefaultAzureMachineRootVolumeSize)
	}
	if obj.ImagefsVolumeEnabled == nil {
		obj.ImagefsVolumeEnabled = pointer.BoolPtr(true)
	}
	if obj.ImagefsVolumeEnabled != nil && *obj.ImagefsVolumeEnabled {
		if obj.ImagefsVolumeType == nil {
			obj.ImagefsVolumeType = pointer.StringPtr(constants.DefaultAzureMachineImagefsVolumeType)
		}
		if obj.ImagefsVolumeSize == nil {
			obj.ImagefsVolumeSize = pointer.Int64Ptr(constants.DefaultAzureMachineImagefsVolumeSize)
		}
	}
}
