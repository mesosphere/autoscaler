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
	"sync"
	"time"

	"github.com/spf13/pflag"

	"github.com/mesosphere/kommander-cluster-lifecycle/pkg/log"
)

type CachedImageMetadataFetcher struct {
	delegateFetcher          ImageVersionGetter
	cacheRefreshInterval     time.Duration
	imageList                []ImageMetadata
	kubernetesVersionToImage map[string]*ImageMetadata
	m                        sync.RWMutex
	logger                   log.Logger
}

type ImageVersionCacheOptions struct {
	*ImageMetadataOptions
	*VersionCacheOptions
}

func (imo *ImageVersionCacheOptions) AddFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	imo.ImageMetadataOptions.AddFlags(flagPrefix, flagSet)
	imo.VersionCacheOptions.AddFlags(flagPrefix, flagSet)
}

func NewImageVersionCacheOptions() *ImageVersionCacheOptions {
	return &ImageVersionCacheOptions{
		ImageMetadataOptions: &ImageMetadataOptions{},
		VersionCacheOptions:  &VersionCacheOptions{},
	}
}

var _ ImageVersioner = &CachedImageMetadataFetcher{}

// NewCachedImageMetadataFetcher creates a CachedImageMetadataFetcher
func NewCachedImageMetadataFetcher(logger log.Logger, delegate ImageVersionGetter, opts VersionCacheOptions) (*CachedImageMetadataFetcher, error) {
	return &CachedImageMetadataFetcher{
		delegateFetcher:      delegate,
		logger:               logger,
		cacheRefreshInterval: opts.CacheRefreshInterval,
	}, nil
}

func (f *CachedImageMetadataFetcher) ImageVersionForKubernetesVersion(kubernetesVersion string) (ImageMetadata, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	im, ok := f.kubernetesVersionToImage[kubernetesVersion]
	if !ok {
		f.logger.WithValues("kubernetesVersion", kubernetesVersion).Info("unable to find the specified Kubernetes version")
		return ImageMetadata{}, fmt.Errorf("unable to find Konvoy version for the specified Kubernetes version %s", kubernetesVersion)
	}
	return *im, nil
}

func (f *CachedImageMetadataFetcher) ListImages() ([]ImageMetadata, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.imageList, nil
}

func (f *CachedImageMetadataFetcher) Start(stop <-chan struct{}) error {
	// Return error if initial caching fails to fail early.
	if err := f.updateCache(); err != nil {
		return err
	}
	if f.cacheRefreshInterval > 0 {
		go func() {
			ticker := time.NewTicker(f.cacheRefreshInterval)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					break
				case <-ticker.C:
					if err := f.updateCache(); err != nil {
						f.logger.Error(err, "failed to update image metadata cache")
					}
				}
			}
		}()
	}
	return nil
}

func (f *CachedImageMetadataFetcher) updateCache() error {
	images, err := f.delegateFetcher.ListImages()
	if err != nil {
		return err
	}
	newCache := make(map[string]*ImageMetadata, len(images))
	for i := range images {
		im := images[i]
		existingKonvoyVersion, ok := newCache[im.DefaultKubernetesVersion.String()]
		if !ok || existingKonvoyVersion.Version.LessThan(im.Version) {
			newCache[im.DefaultKubernetesVersion.String()] = &im
		}
	}

	f.m.Lock()
	defer f.m.Unlock()
	f.imageList = images
	f.kubernetesVersionToImage = newCache
	return nil
}
