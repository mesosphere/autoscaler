package revisions

import (
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
)

// AddonRevisions is a list of historical versions of a single addon
type AddonRevisions []v1beta1.AddonInterface
