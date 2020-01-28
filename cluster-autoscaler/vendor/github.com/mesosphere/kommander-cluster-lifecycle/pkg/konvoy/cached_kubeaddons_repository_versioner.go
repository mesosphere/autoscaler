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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver"
	"github.com/mesosphere/kubeaddons/pkg/constants"
	"github.com/spf13/pflag"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/mesosphere/kommander-cluster-lifecycle/pkg/log"
)

var (
	_ KubeaddonsRepositoryVersioner = &CachedKubeaddonsRepositoryVersioner{}

	// viableTagPrefixes is used to weed out a variety of tags that we wouldn't want to use in our cache.
	// any tag prefixed with `override-` is considered special as we needed the ability to dynamically map
	// tags to a version via the command line arguments.
	// TODO - later we'll need to either remove, or flesh out the overrides present in the helm chart
	viableTagPrefixes = []string{"override-", "stable-"}
)

// KubeaddonsRepositoryVersion represents a kubeaddons-configs version that has all the parts to identify it
type KubeaddonsRepositoryVersion struct {
	Distribution      string
	ReleaseNumber     uint64
	KubernetesVersion semver.Version
}

// Tag returns the tag of the KubeaddonsRepositoryVersion
func (v KubeaddonsRepositoryVersion) Tag() string {
	return strings.Join([]string{v.Distribution, v.KubernetesVersion.String(), strconv.FormatUint(v.ReleaseNumber, 10)}, "-")
}

type CachedKubeaddonsRepositoryVersioner struct {
	kubernetesVersionToKubeaddonsRepositoryVersion map[string]KubeaddonsRepositoryVersion
	m                                              sync.RWMutex
	logger                                         log.Logger
	cacheInterval                                  time.Duration
}

type VersionCacheOptions struct {
	CacheRefreshInterval time.Duration // Set to 0 to disable.
}

func (o *VersionCacheOptions) AddFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	flagSet.DurationVar(
		&o.CacheRefreshInterval,
		flagPrefix+"version-cache-refresh-interval",
		time.Hour,
		fmt.Sprintf("kubeaddons repository version cache refresh interval (only applicable for %s strategy)", KubeaddonsRepositoryKubeaddonsRepositoryTagVersionStrategy),
	)
}

func (f *CachedKubeaddonsRepositoryVersioner) KubeaddonsRepositoryVersionForKubernetesVersion(kubernetesVersion string) (string, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	kubeaddonsConfigVersion, ok := f.kubernetesVersionToKubeaddonsRepositoryVersion[kubernetesVersion]
	if !ok {
		return "", fmt.Errorf("unable to find the specified Kubernetes version")
	}
	return kubeaddonsConfigVersion.Tag(), nil
}

func (f *CachedKubeaddonsRepositoryVersioner) Start(stop <-chan struct{}) error {
	// Return error if initial caching fails to fail early.
	if err := f.updateCache(); err != nil {
		return err
	}
	if f.cacheInterval > 0 {
		go func() {
			ticker := time.NewTicker(f.cacheInterval)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					break
				case <-ticker.C:
					if err := f.updateCache(); err != nil {
						f.logger.Error(err, "failed to update kubeaddons configs cache")
					}
				}
			}
		}()
	}
	return nil
}

func (f *CachedKubeaddonsRepositoryVersioner) updateCache() error {
	// TODO: when repo size gets too big we'll need to tweak the memory limits of the kcl controller pod
	repository, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: constants.DefaultConfigRepo})
	if err != nil {
		return fmt.Errorf("failed to clone repo `%s`: %w", constants.DefaultConfigRepo, err)
	}

	tagIter, err := repository.Tags()
	if err != nil {
		return fmt.Errorf("error getting repository.Tags(): %w", err)
	}

	tag2version := make(map[string]KubeaddonsRepositoryVersion)
	err = tagIter.ForEach(func(reference *plumbing.Reference) error {
		tag := reference.Name().Short()

		viable := false
		for _, viablePrefix := range viableTagPrefixes {
			if strings.HasPrefix(tag, viablePrefix) {
				viable = true
			}
		}

		if !viable {
			return nil
		}

		version, err := parseKubeaddonsRepositoryVersion(tag)
		if err != nil {
			return fmt.Errorf("failed parsing tag `%s`: %w", tag, err)
		}

		existingVersion, ok := tag2version[version.KubernetesVersion.String()]
		if !ok || version.Distribution == "override" || existingVersion.ReleaseNumber < version.ReleaseNumber {
			tag2version[version.KubernetesVersion.String()] = version
		}

		return nil
	})
	if err != nil {
		return err
	}

	f.m.Lock()
	defer f.m.Unlock()

	f.kubernetesVersionToKubeaddonsRepositoryVersion = tag2version

	return nil
}

// parseKubeaddonsRepositoryVersion returns a KubeaddonsRepositoryVersion given the version string
func parseKubeaddonsRepositoryVersion(version string) (KubeaddonsRepositoryVersion, error) {
	parts := strings.Split(version, "-")

	// if a release number was not provided, just assume it is the oldest release
	if len(parts) == 2 {
		version = fmt.Sprintf("%s-0", version)
		parts = strings.Split(version, "-")
	}

	if len(parts) != 3 {
		return KubeaddonsRepositoryVersion{}, fmt.Errorf("version `%s` does not follow standard KubeaddonsRepositoryVersion", version)
	}

	distribution := parts[0]
	if distribution != "stable" {
		return KubeaddonsRepositoryVersion{}, fmt.Errorf("version `%s` does not have a stable distribution", version)
	}
	kubernetesVersion, err := semver.Parse(parts[1])
	if err != nil {
		return KubeaddonsRepositoryVersion{}, fmt.Errorf("version `%s`'s kubernetes version part (%s) does not follow semver: %w", version, parts[1], err)
	}
	releaseNumber, err := strconv.ParseUint(parts[2], 10, 0)
	if err != nil {
		return KubeaddonsRepositoryVersion{}, fmt.Errorf("version `%s`'s kubernetes release number part (%s) is not an unint: %w", version, parts[2], err)
	}

	return KubeaddonsRepositoryVersion{
		Distribution:      distribution,
		KubernetesVersion: kubernetesVersion,
		ReleaseNumber:     releaseNumber,
	}, nil
}
