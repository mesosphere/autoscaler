package repositories

import "github.com/mesosphere/kubeaddons/pkg/api/v1beta1"

// AddonFilter is a function that will return `true` if the addon is
// to be included in the resulting list.
type AddonFilter func(addon v1beta1.AddonInterface) bool
