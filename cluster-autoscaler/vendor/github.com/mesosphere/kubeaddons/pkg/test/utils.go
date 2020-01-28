package test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"

	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	addonerrs "github.com/mesosphere/kubeaddons/pkg/errors"
)

// -----------------------------------------------------------------------------
// Utils - Public Functions
// -----------------------------------------------------------------------------

// DecodeObjectFromManifest decodes Addons and ClusterAddons from yaml source
func DecodeObjectFromManifest(data []byte) (runtime.Object, error) {
	scheme := runtime.NewScheme()
	if err := v1beta1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// apiextensionsv1beta1.AddToScheme(scheme)
	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if runtime.IsNotRegisteredError(err) {
		return nil, errors.New(addonerrs.ErrorDecodedObjectNotAddonOrClusterAddon)
	}
	if runtime.IsMissingKind(err) {
		return nil, errors.New(addonerrs.ErrorDecodedObjectNotAddonOrClusterAddon)
	}
	if err != nil {
		if strings.Contains(err.Error(), "yaml: line ") {
			return nil, errors.New(addonerrs.ErrorDecodedObjectNotAddonOrClusterAddon)
		}
		if strings.Contains(err.Error(), "cannot unmarshal string into Go value of type struct") {
			return nil, errors.New(addonerrs.ErrorDecodedObjectNotAddonOrClusterAddon)
		}
		return nil, err
	}
	return obj, nil
}

func WaitForKubernetes(c kubernetes.Interface, clusterPodsTimeout time.Duration) error {
	timeout := time.Now().Add(clusterPodsTimeout)

	for timeout.After(time.Now()) {
		ds, err := c.AppsV1().Deployments("kube-system").List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failure pulling deployments while waiting for kubernetes: %w", err)
		}
		for _, d := range ds.Items {
			if d.Status.ReadyReplicas < 1 {
				continue
			}
		}

		dms, err := c.AppsV1().DaemonSets("kube-system").List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failure pulling deployments while waiting for kubernetes: %w", err)
		}
		for _, dm := range dms.Items {
			if dm.Status.NumberReady < 1 {
				continue
			}
		}

		ss, err := c.AppsV1().StatefulSets("kube-system").List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failure pulling deployments while waiting for kubernetes: %w", err)
		}
		for _, s := range ss.Items {
			if s.Status.ReadyReplicas < 1 {
				continue
			}
		}

		return nil
	}

	return fmt.Errorf("timed out waiting for kubernetes after %s", clusterPodsTimeout)
}
