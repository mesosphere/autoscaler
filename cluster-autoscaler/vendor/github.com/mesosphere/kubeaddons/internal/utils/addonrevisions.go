package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver"

	"github.com/mesosphere/kubeaddons/pkg/constants"
	"github.com/mesosphere/kubeaddons/pkg/repositories/addons/revisions"
)

// SortAddonRevisions sorts a provided list of Addon resources by their defined revision version
// TODO: later we want to implement the standard Sort interface on the AddonRevisions type
//       as that should be pretty straightforwardbut at the time of writing, we were already
//       using this methodology and could copy it.
func SortAddonRevisions(revs revisions.AddonRevisions) (err error) {
	sort.Slice(revs, func(i, j int) bool {
		rev1, ok := revs[i].GetAnnotations()[constants.AddonRevisionAnnotation]
		if !ok {
			if err == nil {
				err = fmt.Errorf("addon %s does not have a valid addon-revision", revs[i].GetName())
			} else {
				err = fmt.Errorf("addon %s does not have a valid addon-revision: %w", revs[i].GetName(), err)
			}
		}

		rev2 := revs[j].GetAnnotations()[constants.AddonRevisionAnnotation]
		if !ok {
			if err == nil {
				err = fmt.Errorf("addon %s does not have a valid addon-revision", revs[i].GetName())
			} else {
				err = fmt.Errorf("addon %s does not have a valid addon-revision: %w", revs[i].GetName(), err)
			}
		}

		if rev1 == "" {
			rev1 = "0.0.0-1"
		}

		if rev2 == "" {
			rev2 = "0.0.0-1"
		}

		v1 := semver.MustParse(strings.TrimPrefix(rev1, "v"))
		v2 := semver.MustParse(strings.TrimPrefix(rev2, "v"))

		return v1.LT(v2)
	})

	// reverse - latest addon revision first
	for i, j := 0, len(revs)-1; i < j; i, j = i+1, j-1 {
		revs[i], revs[j] = revs[j], revs[i]
	}

	return
}
