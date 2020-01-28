package repositories

import (
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
)

// -----------------------------------------------------------------------------
// Public Functions
// -----------------------------------------------------------------------------

// IncludeAddon consumes a list of AddonFilters and a provided addon and returns boolean
// whether or not the addon should be included based on those filters.
func IncludeAddon(addon v1beta1.AddonInterface, filters []AddonFilter) bool {
	for _, filter := range filters {
		if !filter(addon) {
			return false
		}
	}
	return true
}
