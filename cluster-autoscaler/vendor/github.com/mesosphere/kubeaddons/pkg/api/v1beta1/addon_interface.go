package v1beta1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +k8s:deepcopy-gen=false
// AddonInterface is an interface that represents a Kubernetes Addon
type AddonInterface interface {
	metav1.Object
	runtime.Object

	// GetStatus returns the status of the Addon
	GetStatus() *AddonStatus
	// GetAddonSpec returns the spec of an Addon, common to all addons
	GetAddonSpec() *AddonSpec
	// DeployNamespace returns the namespace to deploy the addon
	DeployNamespace() string

	// NewEmptyType returns an AddonInterface of the same internal type that's empty
	NewEmptyType() AddonInterface
}

// GetDeps returns all the addon dependencies given the LabelSelector
func GetDeps(c client.Client, label *metav1.LabelSelector) ([]AddonInterface, error) {
	lsmap, err := metav1.LabelSelectorAsMap(label)
	if err != nil {
		return nil, fmt.Errorf("failure processing label selector: %w", err)
	}

	deps := []AddonInterface{}

	addonDeps := &AddonList{}
	if err := c.List(context.Background(), addonDeps, client.MatchingLabels(lsmap)); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting addon deps: %w", err)
		}
	}
	for _, addon := range addonDeps.Items {
		deps = append(deps, &addon)
	}

	clusterAddonDeps := &ClusterAddonList{}
	if err := c.List(context.Background(), clusterAddonDeps, client.MatchingLabels(lsmap)); err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("error getting clusteraddon deps: %w", err)
		}
	}
	for _, clusterAddon := range clusterAddonDeps.Items {
		deps = append(deps, &clusterAddon)
	}

	return deps, nil
}
