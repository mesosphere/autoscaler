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
	"github.com/mesosphere/kubernetes-resource-manager/pkg/addons"

	"github.com/mesosphere/kommander-cluster-lifecycle/pkg/log"
)

type DefaultKubeaddonsGetter interface {
	DefaultAddonsForRepoAndProvider(addonsConfigRepository, addonsConfigVersion, provider string) addons.AddonConfigs
}

type DirectDefaultKubeaddonsGetter struct {
	Logger log.Logger
}

var _ DefaultKubeaddonsGetter = &DirectDefaultKubeaddonsGetter{}

func (d *DirectDefaultKubeaddonsGetter) DefaultAddonsForRepoAndProvider(addonsConfigRepository, addonsConfigVersion, provider string) addons.AddonConfigs {
	configs, err := addons.GetDefaultAddonConfigs(
		provider,
		addons.TemplateRepos{{
			URL: addonsConfigRepository,
			Tag: addonsConfigVersion,
		}},
	)
	if err != nil {
		d.Logger.Error(err, "failed to get default addon configs", "addonsRepository", addonsConfigRepository, "version", addonsConfigVersion, "provider", provider)
		return nil
	}
	return configs
}
