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
 */

package konvoy

import (
	"bytes"
	"fmt"

	"github.com/mesosphere/konvoy/pkg/apis/konvoy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	lineSeparator     = "\n"
	yamlFileSeparator = "---"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	v1beta1.AddToScheme(scheme)
}

func ConvertKonvoyAPIConfigToYAML(
	clusterName string,
	clusterCreationTimestamp metav1.Time,
	provisionerSpec v1beta1.ClusterProvisionerSpec,
	clusterSpec v1beta1.ClusterConfigurationSpec,
) (string, error) {
	var buf bytes.Buffer
	yamlSerializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory, scheme, scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: false,
			Strict: false,
		},
	)

	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(yamlSerializer, v1beta1.GroupVersion)

	provisionerConfig := &v1beta1.ClusterProvisioner{
		ObjectMeta: metav1.ObjectMeta{
			Name:              clusterName,
			CreationTimestamp: clusterCreationTimestamp,
		},
		Spec: provisionerSpec,
	}

	buf.WriteString(yamlFileSeparator)
	buf.WriteString(lineSeparator)

	if err := encoder.Encode(provisionerConfig, &buf); err != nil {
		return "", fmt.Errorf("failed to encode Konvoy provisioner config: %v", err)
	}

	clusterConfig := &v1beta1.ClusterConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:              clusterName,
			CreationTimestamp: clusterCreationTimestamp,
		},
		Spec: clusterSpec,
	}

	buf.WriteString(lineSeparator)
	buf.WriteString(yamlFileSeparator)
	buf.WriteString(lineSeparator)

	if err := encoder.Encode(clusterConfig, &buf); err != nil {
		return "", fmt.Errorf("failed to encode Konvoy cluster config: %v", err)
	}

	return buf.String(), nil
}
