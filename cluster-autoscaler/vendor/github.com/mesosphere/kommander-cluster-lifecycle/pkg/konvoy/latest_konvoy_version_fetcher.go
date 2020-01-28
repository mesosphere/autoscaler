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

type LatestKonvoyImageVersionGetter struct {
	delegateFetcher ImageVersionGetter
}

func NewLatestImageVersionGetter(delegateFetcher ImageVersionGetter) ImageVersioner {
	return &LatestKonvoyImageVersionGetter{delegateFetcher: delegateFetcher}
}

func (v *LatestKonvoyImageVersionGetter) ImageVersionForKubernetesVersion(_ string) (ImageMetadata, error) {
	return v.latestKonvoyImageVersion()
}

func (v *LatestKonvoyImageVersionGetter) ListImages() ([]ImageMetadata, error) {
	latestImage, err := v.latestKonvoyImageVersion()
	if err != nil {
		return nil, err
	}
	return []ImageMetadata{latestImage}, nil
}

func (v *LatestKonvoyImageVersionGetter) latestKonvoyImageVersion() (ImageMetadata, error) {
	images, err := v.delegateFetcher.ListImages()
	if err != nil {
		return ImageMetadata{}, err
	}
	if len(images) > 0 {
		return images[0], nil
	}
	return ImageMetadata{}, nil
}

func (v *LatestKonvoyImageVersionGetter) Start(stop <-chan struct{}) error {
	if kiv, ok := v.delegateFetcher.(ImageVersioner); ok {
		return kiv.Start(stop)
	}
	return nil
}
