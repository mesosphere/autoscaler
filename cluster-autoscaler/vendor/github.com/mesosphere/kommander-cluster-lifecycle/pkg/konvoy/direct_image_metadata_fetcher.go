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
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/docker-ce/components/engine/image"
	"github.com/mesosphere/konvoy/pkg/constants"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/pflag"
	utilversion "k8s.io/apimachinery/pkg/util/version"

	"github.com/mesosphere/kommander-cluster-lifecycle/pkg/log"
)

const (
	VersionLabelName                    = "KONVOY_VERSION"
	DefaultKubernetesVersionLabelName   = "DEFAULT_KUBERNETES_VERSION"
	MinSupportedWrapperVersionLabelName = "MIN_SUPPORTED_WRAPPER_VERSION"
)

type DirectImageMetadataFetcher struct {
	repository              distribution.Repository
	allowUnofficialReleases bool
	logger                  log.Logger
}

var _ ImageVersionGetter = &DirectImageMetadataFetcher{}

// NewDirectImageMetadataFetcher creates a DirectImageMetadataFetcher
func NewDirectImageMetadataFetcher(opts ImageMetadataOptions, logger log.Logger) (*DirectImageMetadataFetcher, error) {
	r, err := connectToRegistry(opts)
	if err != nil {
		return nil, err
	}
	km := &DirectImageMetadataFetcher{
		repository:              r,
		allowUnofficialReleases: opts.AllowUnofficialReleases,
		logger:                  logger,
	}
	return km, nil
}

// ImageVersionForKubernetesVersion returns a Konvoy version with its metadata for the specific
// Kubernetes version.
func (k *DirectImageMetadataFetcher) ImageVersionForKubernetesVersion(kubernetesVersion string) (ImageMetadata, error) {
	semanticKubernetesVersion, err := utilversion.ParseSemantic(kubernetesVersion)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("requested Kubernetes version %s is not a valid semantic version: %w", kubernetesVersion, err)
	}
	ctx := context.Background()

	tagService := k.repository.Tags(ctx)
	tags, err := tagService.All(ctx)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("unable to list tags the docker repository: %v", err)
	}
	blobs := k.repository.Blobs(ctx)
	ms, err := k.repository.Manifests(ctx)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("unable to create the manifest service: %v", err)
	}

	for _, tagName := range tags {
		var contentDigest digest.Digest
		manifest, err := ms.Get(ctx, "", distribution.WithTag(tagName), distribution.WithManifestMediaTypes([]string{schema2.MediaTypeManifest}), client.ReturnContentDigest(&contentDigest))
		if err != nil {
			k.logger.Info("unable to get the manifest for tag - skipping", "tag", tagName, "error", err)
			continue
		}
		// we can safely assume it is a schema2
		schema2Manifest, _ := manifest.(*schema2.DeserializedManifest)
		targetDescriptor := schema2Manifest.Target()

		blobBytes, err := blobs.Get(ctx, targetDescriptor.Digest)
		if err != nil {
			k.logger.Info("unable to get blob for tag - skipping", "tag", tagName, "digest", targetDescriptor.Digest, "error", err)
			continue
		}

		var image image.V1Image
		if err := json.Unmarshal(blobBytes, &image); err != nil {
			k.logger.Info("unable to unmarshal Konvoy Docker image manifest", "error", err)
			continue
		}
		version, err := utilversion.ParseSemantic(image.Config.Labels[VersionLabelName])
		if err != nil {
			k.logger.Info("invalid semantic version for Konvoy - skipping image", "konvoyVersion", image.Config.Labels[VersionLabelName], "error", err)
			continue
		}
		defaultKubernetesVersion, err := utilversion.ParseSemantic(image.Config.Labels[DefaultKubernetesVersionLabelName])
		if err != nil {
			k.logger.Info("invalid semantic version for default Kubernetes version - skipping image", "kubernetesVersion", image.Config.Labels[DefaultKubernetesVersionLabelName], "error", err)
			continue
		}
		minSupportedWrapperVersion, err := utilversion.ParseSemantic(image.Config.Labels[MinSupportedWrapperVersionLabelName])
		if err != nil {
			k.logger.Info(
				"invalid semantic version for minimum supported wrapper version - skipping image",
				"minSupportedWrapperVersion", image.Config.Labels[MinSupportedWrapperVersionLabelName],
				"error", err,
			)
			continue
		}
		im := ImageMetadata{
			Version:                    version,
			DefaultKubernetesVersion:   defaultKubernetesVersion,
			MinSupportedWrapperVersion: minSupportedWrapperVersion,
		}

		// TODO: Fetch the newest konvoy cli
		if semanticKubernetesVersion.String() == im.DefaultKubernetesVersion.String() && (k.allowUnofficialReleases || isOfficialRelease(im.Version)) {
			return im, nil
		}
	}
	return ImageMetadata{}, fmt.Errorf("unable to find the specified Kubernetes version")
}

// ListImages gets the list of image list from which a cli can be upgraded to
func (k *DirectImageMetadataFetcher) ListImages() ([]ImageMetadata, error) {
	ctx := context.Background()

	tagService := k.repository.Tags(ctx)
	tags, err := tagService.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list tags the docker repository: %v", err)
	}
	blobs := k.repository.Blobs(ctx)
	ms, err := k.repository.Manifests(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create the manifest service: %v", err)
	}

	var images []ImageMetadata
	for _, tagName := range tags {
		var contentDigest digest.Digest
		manifest, err := ms.Get(ctx, "", distribution.WithTag(tagName), distribution.WithManifestMediaTypes([]string{schema2.MediaTypeManifest}), client.ReturnContentDigest(&contentDigest))
		if err != nil {
			k.logger.Info("unable to get the manifest for tag - skipping", "tag", tagName, "error", err)
			continue
		}
		// we can safely assume it is a schema2
		schema2Manifest, _ := manifest.(*schema2.DeserializedManifest)
		targetDescriptor := schema2Manifest.Target()

		blobBytes, err := blobs.Get(ctx, targetDescriptor.Digest)
		if err != nil {
			k.logger.Info("unable to get blob for tag - skipping", "tag", tagName, "digest", targetDescriptor.Digest, "error", err)
			continue
		}

		im, err := k.parseImageToMetadata(blobBytes)
		if err != nil {
			k.logger.Info("failed to parse image metadata - skipping", "tag", tagName, "digest", targetDescriptor.Digest, "error", err)
			continue
		}

		if im.MinSupportedWrapperVersion != nil && (k.allowUnofficialReleases || isOfficialRelease(im.Version)) {
			images = append(images, im)
		}
	}

	// Sort the result slice in descending order.
	sort.SliceStable(images, func(i, j int) bool {
		return images[j].Version.LessThan(images[i].Version)
	})

	return images, nil
}

func (k *DirectImageMetadataFetcher) parseImageToMetadata(blobBytes []byte) (ImageMetadata, error) {
	var image image.V1Image
	if err := json.Unmarshal(blobBytes, &image); err != nil {
		return ImageMetadata{}, fmt.Errorf("unable to convert the image: %v", err)
	}
	var (
		konvoyVersion, defaultKubernetesVersion, minSupportedWrapperVersion *utilversion.Version
		err                                                                 error
	)
	value, ok := image.Config.Labels[VersionLabelName]
	if ok {
		konvoyVersion, err = utilversion.ParseSemantic(value)
		// Skip non-semantic version compatible versions.
		if err != nil {
			return ImageMetadata{}, fmt.Errorf("invalid semantic version %s for Konvoy: %w", value, err)
		}
	}
	value, ok = image.Config.Labels[DefaultKubernetesVersionLabelName]
	if ok {
		defaultKubernetesVersion, err = utilversion.ParseSemantic(value)
		if err != nil {
			return ImageMetadata{}, fmt.Errorf("invalid semantic version %s for default Kubernetes version: %w", value, err)
		}
	}
	value, ok = image.Config.Labels[MinSupportedWrapperVersionLabelName]
	if ok {
		minSupportedWrapperVersion, err = utilversion.ParseSemantic(value)
		if err != nil {
			return ImageMetadata{}, fmt.Errorf("invalid semantic version %s for minimum supported wrapper version: %w", image.Config.Labels[MinSupportedWrapperVersionLabelName], err)
		}
	}
	im := ImageMetadata{
		Version:                    konvoyVersion,
		DefaultKubernetesVersion:   defaultKubernetesVersion,
		MinSupportedWrapperVersion: minSupportedWrapperVersion,
	}
	return im, nil
}

func connectToRegistry(opts ImageMetadataOptions) (distribution.Repository, error) {
	var repo distribution.Repository
	transport, err := NewTokenAuthTransport(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of authorization challenges for the docker registry: %v", err)
	}

	repoName, _ := reference.WithName(opts.DockerRegistryRepository)
	dockerRegistryURL := constants.DefaultDockerRegistryURL
	if opts.DockerRegistryURL != "" {
		dockerRegistryURL = opts.DockerRegistryURL
	}

	repo, err = client.NewRepository(repoName, dockerRegistryURL, transport)
	if err != nil {
		return repo, fmt.Errorf("unable to reach the docker repository: %v", err)
	}
	return repo, nil
}

func isOfficialRelease(version *utilversion.Version) bool {
	return version.PreRelease() == "" && version.BuildMetadata() == ""
}

// ImageMetadataOptions object to store all the properties to interact with a docker registry
type ImageMetadataOptions struct {
	DockerRegistryURL            string
	DockerRegistryUsername       string
	DockerRegistryPassword       string
	DockerRegistryAuthSkipVerify bool
	DockerRegistryRepository     string
	AllowUnofficialReleases      bool
}

func (imo *ImageMetadataOptions) AddFlags(flagPrefix string, flagSet *pflag.FlagSet) {
	flagSet.StringVar(&imo.DockerRegistryURL, flagPrefix+"docker-registry-url", constants.DefaultDockerRegistryURL, "Docker registry URL")
	flagSet.StringVar(&imo.DockerRegistryUsername, flagPrefix+"docker-registry-username", "", "Docker registry username")
	flagSet.StringVar(&imo.DockerRegistryPassword, flagPrefix+"docker-registry-password", "", "Docker registry password")

	flagSet.BoolVar(&imo.DockerRegistryAuthSkipVerify, flagPrefix+"docker-registry-insecure-skip-tls-verify", false, "Flag to turn off the Docker registry certificate verification")
	flagSet.StringVar(&imo.DockerRegistryRepository, flagPrefix+"docker-registry-repository", constants.KonvoyRepoName, "Docker registry image repository")
	flagSet.BoolVar(&imo.AllowUnofficialReleases, flagPrefix+"allow-unofficial-releases", false, "Allow unofficial releases")
}

// ImageMetadata object containing the properties of the konvoy image metadata
type ImageMetadata struct {
	// Version konvoy version
	Version *utilversion.Version `json:"konvoy_version" description:"konvoy version"`
	// DefaultKubernetesVersion sets the default Kubernetes version
	DefaultKubernetesVersion *utilversion.Version `json:"kubernetes_version" description:"kubernetes version"`
	// MinSupportedWrapperVersion sets the minimum supported wrapper version to
	// upgrade from a running CLI to a different Konvoy version
	MinSupportedWrapperVersion *utilversion.Version `json:"min_supported_version" description:"minimum version supported to upgrade to"`
}
