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
	"strings"

	"github.com/mesosphere/konvoy/pkg/apis/konvoy/v1beta1"
	kubeaddonsconstants "github.com/mesosphere/kubeaddons/pkg/constants"
	"k8s.io/utils/pointer"

	"github.com/mesosphere/kommander-cluster-lifecycle/pkg/konvoy"
)

// +k8s:defaulter-gen=covers
func SetDefaults_KonvoyManagementClusterSpec(obj *KonvoyManagementClusterSpec) { //nolint:stylecheck
	v1beta1.SetDefaults_ClusterProvisionerSpec(&obj.ProvisionerConfiguration)
	v1beta1.SetDefaultsClusterConfigurationSpecForProvider(obj.ProvisionerConfiguration.Provider, &obj.ClusterConfiguration)

	// Create a temporary ClusterConfiguration to use Konvoy defaults which target the top level type.
	tempClusterConfiguration := v1beta1.ClusterConfiguration{Spec: obj.ClusterConfiguration}
	v1beta1.SetObjectDefaults_ClusterConfiguration(&tempClusterConfiguration)
	// Use the now defaulted ClusterConfigurationSpec.
	obj.ClusterConfiguration = tempClusterConfiguration.Spec

	if obj.UpgradeStrategy == nil {
		obj.UpgradeStrategy = &defaultUpgradeStrategy
	}

	// Set the SSH key paths if they are not already set.
	if obj.ProvisionerConfiguration.SSHCredentials == nil {
		obj.ProvisionerConfiguration.SSHCredentials = &v1beta1.SSHCredentials{}
	}
	if obj.ProvisionerConfiguration.SSHCredentials.PublicKeyFile == "" {
		obj.ProvisionerConfiguration.SSHCredentials.PublicKeyFile = "/tmp/.ssh/ssh-publickey"
	}
	if obj.ProvisionerConfiguration.SSHCredentials.PrivateKeyFile == nil || *obj.ProvisionerConfiguration.SSHCredentials.PrivateKeyFile == "" {
		obj.ProvisionerConfiguration.SSHCredentials.PrivateKeyFile = pointer.StringPtr("/tmp/.ssh/ssh-privatekey")
	}
}

// +kubebuilder:object:generate=false
type KonvoyManagementClusterDefaulterFunc func(*KonvoyManagementCluster)

func ChainedKonvoyManagementClusterDefaulterFunc(dfs ...KonvoyManagementClusterDefaulterFunc) GenericDefaulterFunc {
	return func(obj interface{}) {
		kc := obj.(*KonvoyManagementCluster)
		for _, df := range dfs {
			df(kc)
		}
	}
}

func SetObjectDefaultsFunc_KonvoyManagementCluster_WithKonvoyImageMetadata(imf konvoy.ImageVersionGetter) KonvoyManagementClusterDefaulterFunc { //nolint:stylecheck
	return func(obj *KonvoyManagementCluster) {
		kubernetes := obj.Spec.ClusterConfiguration.Kubernetes
		// We can't default anything if we don't know the Kubernetes version.
		if kubernetes == nil || kubernetes.Version == nil {
			return
		}
		konvoyImageMetadata, err := imf.ImageVersionForKubernetesVersion(*kubernetes.Version)
		// An error in this case can mean that no image has been found for this Kubernetes version so
		// we cannot default anything else here.
		if err != nil {
			return
		}
		konvoyVersion := konvoyImageMetadata.Version.String()
		if !strings.HasPrefix(konvoyVersion, "v") {
			konvoyVersion = "v" + konvoyVersion
		}
		if obj.Spec.ClusterConfiguration.Version == nil || *obj.Spec.ClusterConfiguration.Version == "" {
			obj.Spec.ClusterConfiguration.Version = pointer.StringPtr(konvoyVersion)
		}
		if obj.Spec.ProvisionerConfiguration.Version == nil || *obj.Spec.ProvisionerConfiguration.Version == "" {
			obj.Spec.ProvisionerConfiguration.Version = pointer.StringPtr(konvoyVersion)
		}
	}
}

func SetObjectDefaultsFunc_KonvoyManagementCluster_WithKubeaddonsRepository(kacf konvoy.KubeaddonsRepositoryVersioner) KonvoyManagementClusterDefaulterFunc { //nolint:stylecheck
	return func(obj *KonvoyManagementCluster) {
		kubernetes := obj.Spec.ClusterConfiguration.Kubernetes
		// We can't default anything if we don't know the Kubernetes version.
		if kubernetes == nil || kubernetes.Version == nil {
			return
		}
		// Get the default kubeddons-configs version for this version of Konvoy.
		kubeaddonsConfigVersion, err := kacf.KubeaddonsRepositoryVersionForKubernetesVersion(*kubernetes.Version)
		if err != nil {
			return
		}

		if len(obj.Spec.ClusterConfiguration.Addons) == 0 {
			obj.Spec.ClusterConfiguration.Addons = []v1beta1.Addons{{}}
		}
		for i := range obj.Spec.ClusterConfiguration.Addons {
			addons := &obj.Spec.ClusterConfiguration.Addons[i]
			// Don't set the version if already set.
			if addons.ConfigVersion != nil && *addons.ConfigVersion != "" {
				continue
			}
			// Only set the version if this is for the kubeaddons repository repository.
			if addons.ConfigRepository == nil || *addons.ConfigRepository == "" {
				addons.ConfigRepository = pointer.StringPtr(kubeaddonsconstants.DefaultConfigRepo)
			}
			if *addons.ConfigRepository == kubeaddonsconstants.DefaultConfigRepo {
				addons.ConfigVersion = pointer.StringPtr(kubeaddonsConfigVersion)
			}
		}
	}
}

func SetObjectDefaultsFunc_KonvoyManagementCluster_DefaultKubeaddonsRepository(kag konvoy.DefaultKubeaddonsGetter) KonvoyManagementClusterDefaulterFunc { //nolint:stylecheck
	return func(obj *KonvoyManagementCluster) {
		for i := range obj.Spec.ClusterConfiguration.Addons {
			addons := &obj.Spec.ClusterConfiguration.Addons[i]
			if len(addons.AddonList) > 0 ||
				addons.ConfigRepository == nil || *addons.ConfigRepository != kubeaddonsconstants.DefaultConfigRepo ||
				addons.ConfigVersion == nil || *addons.ConfigVersion == "" {
				continue
			}

			addonConfigs := kag.DefaultAddonsForRepoAndProvider(*addons.ConfigRepository, *addons.ConfigVersion, obj.Spec.ProvisionerConfiguration.Provider)
			if addonConfigs != nil {
				// Disable kommander addon if it is enabled, otherwise there is a race if kommander addon
				// gets deployed on created clusters.
				disableAddonIfPresent(KommanderAddonName, addonConfigs)
				addons.AddonList = addonConfigs
			}
		}
	}
}
