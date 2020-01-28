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

package konvoy

import (
	"fmt"

	"github.com/spf13/pflag"
)

var (
	_ KubeaddonsRepositoryVersioner = &MappedKubeaddonsRepositoryVersioner{}
)

type MappedKubeaddonsRepositoryVersioner struct {
	kubernetesVersionToKubeaddonsRepositoryVersion map[string]string
}

type KubeaddonsRepositoryVersionMappingOptions struct {
	KubernetesVersionToKubeaddonsRepositoryVersion map[string]string
}

func (o *KubeaddonsRepositoryVersionMappingOptions) AddFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	flagSet.StringToStringVar(
		&o.KubernetesVersionToKubeaddonsRepositoryVersion,
		flagPrefix+"version-map",
		nil,
		fmt.Sprintf("Kubernetes version to kubeaddons repository version map (only applicable for %s strategy)", MappedKubeaddonsRepositoryVersionStrategy))
}

func (f *MappedKubeaddonsRepositoryVersioner) KubeaddonsRepositoryVersionForKubernetesVersion(kubernetesVersion string) (string, error) {
	kubeaddonsConfigVersion, ok := f.kubernetesVersionToKubeaddonsRepositoryVersion[kubernetesVersion]
	if !ok {
		return "", fmt.Errorf("unable to find the specified Kubernetes version")
	}
	return kubeaddonsConfigVersion, nil
}

func (f *MappedKubeaddonsRepositoryVersioner) Start(stop <-chan struct{}) error {
	return nil
}
