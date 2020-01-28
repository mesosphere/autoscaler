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

type ImageVersionStrategy string

func newImageVersionStrategyValue(val ImageVersionStrategy, p *ImageVersionStrategy) *ImageVersionStrategy {
	*p = val
	return p
}

func (s *ImageVersionStrategy) Set(v string) error {
	strategy := ImageVersionStrategy(v)
	switch strategy {
	case DockerImageLabelsImageVersionStrategy, LatestReleaseImageVersionStrategy:
		*s = strategy
		return nil
	default:
		return fmt.Errorf(
			"invalid value: %s. Must be one of: %v",
			v,
			imageVersionStrategyChoices,
		)
	}
}

func (s *ImageVersionStrategy) Type() string {
	return "ImageVersionStrategy"
}

func (s *ImageVersionStrategy) String() string {
	return string(*s)
}

const (
	DockerImageLabelsImageVersionStrategy ImageVersionStrategy = "docker-image-labels"
	LatestReleaseImageVersionStrategy     ImageVersionStrategy = "latest"
)

var imageVersionStrategyChoices = []ImageVersionStrategy{
	DockerImageLabelsImageVersionStrategy,
	LatestReleaseImageVersionStrategy,
}

type ImageVersionGetter interface {
	ImageVersionForKubernetesVersion(kubernetesVersion string) (ImageMetadata, error)
	ListImages() ([]ImageMetadata, error)
}

type ImageVersioner interface {
	ImageVersionGetter
	Start(<-chan struct{}) error
}

type ImageVersionOptions struct {
	*ImageVersionCacheOptions

	ImageVersionStrategy
}

func NewImageVersionOptions() *ImageVersionOptions {
	return &ImageVersionOptions{
		ImageVersionCacheOptions: NewImageVersionCacheOptions(),
	}
}

func (o *ImageVersionOptions) AddFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	o.ImageVersionCacheOptions.AddFlags(flagPrefix, flagSet)

	flagSet.Var(
		newImageVersionStrategyValue(LatestReleaseImageVersionStrategy, &o.ImageVersionStrategy),
		flagPrefix+"version-strategy", fmt.Sprintf("konvoy version strategy (one of: %v)", imageVersionStrategyChoices),
	)
}

func NewImageVersioner(opts ImageVersionOptions, logger log.Logger) (ImageVersioner, error) {
	switch opts.ImageVersionStrategy {
	case DockerImageLabelsImageVersionStrategy:
		labelVersionGetter, err := NewDirectImageMetadataFetcher(*opts.ImageVersionCacheOptions.ImageMetadataOptions, logger.WithName("direct"))
		if err != nil {
			return nil, err
		}
		return NewCachedImageMetadataFetcher(logger, labelVersionGetter, *opts.ImageVersionCacheOptions.VersionCacheOptions)
	case LatestReleaseImageVersionStrategy:
		labelVersionGetter, err := NewDirectImageMetadataFetcher(*opts.ImageVersionCacheOptions.ImageMetadataOptions, logger.WithName("direct"))
		if err != nil {
			return nil, err
		}
		cachedFetcher, err := NewCachedImageMetadataFetcher(logger, labelVersionGetter, *opts.ImageVersionCacheOptions.VersionCacheOptions)
		if err != nil {
			return nil, err
		}
		return NewLatestImageVersionGetter(cachedFetcher), nil
	default:
		return nil,
			fmt.Errorf(
				"unknown kubeaddons repository version strategy: %s. Must be one of: %v",
				opts.ImageVersionStrategy,
				imageVersionStrategyChoices,
			)
	}
}
