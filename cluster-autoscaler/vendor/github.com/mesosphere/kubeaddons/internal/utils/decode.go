package utils

import (
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/errors"
	"github.com/mesosphere/kubeaddons/pkg/test"
)

// -----------------------------------------------------------------------------
// Public Functions
// -----------------------------------------------------------------------------

// IsThisAnAddon provides boolean whether or not a given set of bytes is a yaml
// addon object manifest.
func IsThisAnAddon(fileContents []byte) (bool, v1beta1.AddonInterface, error) {
	runtimeObj, err := test.DecodeObjectFromManifest(fileContents)
	if err != nil {
		if err.Error() != errors.ErrorDecodedObjectNotAddonOrClusterAddon {
			return false, nil, nil
		}
		return false, nil, err
	}

	gvk := runtimeObj.GetObjectKind().GroupVersionKind()
	if gvk.Kind == "Addon" || gvk.Kind == "ClusterAddon" {
		if addon, ok := runtimeObj.(v1beta1.AddonInterface); ok {
			return true, addon, nil
		}
	}

	return false, nil, nil
}
