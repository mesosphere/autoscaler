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

	"github.com/mesosphere/kommander-cluster-lifecycle/pkg/log"
)

type KubeaddonsRepositoryVersioner interface {
	KubeaddonsRepositoryVersionForKubernetesVersion(kubernetesVersion string) (string, error)
	Start(<-chan struct{}) error
}

type KubeaddonsRepositoryVersionOptions struct {
	*VersionCacheOptions
	*KubeaddonsRepositoryVersionMappingOptions

	KubeaddonsRepositoryVersionStrategy
}

type KubeaddonsRepositoryVersionStrategy string

func newKubeaddonsRepositoryVersionStrategyValue(val KubeaddonsRepositoryVersionStrategy, p *KubeaddonsRepositoryVersionStrategy) *KubeaddonsRepositoryVersionStrategy {
	*p = val
	return p
}

func (s *KubeaddonsRepositoryVersionStrategy) Set(v string) error {
	strategy := KubeaddonsRepositoryVersionStrategy(v)
	switch strategy {
	case MasterKubeaddonsRepositoryVersionStrategy, KubeaddonsRepositoryKubeaddonsRepositoryTagVersionStrategy, MappedKubeaddonsRepositoryVersionStrategy:
		*s = strategy
		return nil
	default:
		return fmt.Errorf(
			"invalid value: %s. Must be one of: %v",
			v,
			kubeaddonsRepositoryVersionStrategyChoices,
		)
	}
}

func (s *KubeaddonsRepositoryVersionStrategy) Type() string {
	return "KubeaddonsRepositoryVersionStrategy"
}

func (s *KubeaddonsRepositoryVersionStrategy) String() string {
	return string(*s)
}

const (
	KubeaddonsRepositoryKubeaddonsRepositoryTagVersionStrategy KubeaddonsRepositoryVersionStrategy = "repository-tag"
	MasterKubeaddonsRepositoryVersionStrategy                  KubeaddonsRepositoryVersionStrategy = "master"
	MappedKubeaddonsRepositoryVersionStrategy                  KubeaddonsRepositoryVersionStrategy = "mapped-kubernetes-version"
)

var kubeaddonsRepositoryVersionStrategyChoices = []KubeaddonsRepositoryVersionStrategy{
	KubeaddonsRepositoryKubeaddonsRepositoryTagVersionStrategy,
	MasterKubeaddonsRepositoryVersionStrategy,
	MappedKubeaddonsRepositoryVersionStrategy,
}

func (o *KubeaddonsRepositoryVersionOptions) AddFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	flagSet.Var(
		newKubeaddonsRepositoryVersionStrategyValue(KubeaddonsRepositoryKubeaddonsRepositoryTagVersionStrategy, &o.KubeaddonsRepositoryVersionStrategy),
		flagPrefix+"version-strategy", fmt.Sprintf("kubeaddons repository version strategy (one of: %v)", kubeaddonsRepositoryVersionStrategyChoices),
	)

	o.VersionCacheOptions.AddFlags(flagPrefix, flagSet)
	o.KubeaddonsRepositoryVersionMappingOptions.AddFlags(flagPrefix, flagSet)
}

func NewKubeaddonsRepositoryVersionOptions() *KubeaddonsRepositoryVersionOptions {
	return &KubeaddonsRepositoryVersionOptions{
		VersionCacheOptions:                       &VersionCacheOptions{},
		KubeaddonsRepositoryVersionMappingOptions: &KubeaddonsRepositoryVersionMappingOptions{},
	}
}

func NewKubeaddonsRepositoryVersionFetcher(opts KubeaddonsRepositoryVersionOptions, logger log.Logger) (KubeaddonsRepositoryVersioner, error) {
	switch opts.KubeaddonsRepositoryVersionStrategy {
	case MasterKubeaddonsRepositoryVersionStrategy:
		return &MasterKubeaddonsRepositoryVersioner{}, nil
	case KubeaddonsRepositoryKubeaddonsRepositoryTagVersionStrategy:
		return &CachedKubeaddonsRepositoryVersioner{
			logger:        logger,
			cacheInterval: opts.VersionCacheOptions.CacheRefreshInterval,
		}, nil
	case MappedKubeaddonsRepositoryVersionStrategy:
		return &MappedKubeaddonsRepositoryVersioner{
			kubernetesVersionToKubeaddonsRepositoryVersion: opts.KubeaddonsRepositoryVersionMappingOptions.KubernetesVersionToKubeaddonsRepositoryVersion,
		}, nil
	default:
		return nil,
			fmt.Errorf(
				"unknown kubeaddons repository version strategy: %s. Must be one of: %v",
				opts.KubeaddonsRepositoryVersionStrategy,
				kubeaddonsRepositoryVersionStrategyChoices,
			)
	}
}
