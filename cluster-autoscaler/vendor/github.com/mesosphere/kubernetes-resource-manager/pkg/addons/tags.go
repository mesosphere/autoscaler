package addons

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
)

// KubeaddonsConfigsVersion represents a kubeaddons-configs version that has all the parts to identify it
type KubeaddonsConfigsVersion struct {
	Distribution      KubeaddonsConfigsDistribution
	KubernetesVersion semver.Version
	ReleaseNumber     uint64
}

const tagSeparator = "-"

// Tag returns the tag of the KubeaddonsConfigsVersion
func (v KubeaddonsConfigsVersion) Tag() string {
	return strings.Join([]string{string(v.Distribution), v.KubernetesVersion.String(),
		strconv.FormatUint(v.ReleaseNumber, 10)}, tagSeparator)
}

// KubeaddonsConfigsDistribution is the distribution type of the KubeaddonsConfigsVersion
type KubeaddonsConfigsDistribution string

// StableKubeaddonsConfigsDistribution is the stable distribution type
const StableKubeaddonsConfigsDistribution KubeaddonsConfigsDistribution = "stable"

// ParseKubeaddonsConfigsVersion returns a KubeaddonsConfigsVersion given the version string
func ParseKubeaddonsConfigsVersion(version string) (KubeaddonsConfigsVersion, error) {
	parts := strings.Split(version, "-")
	if len(parts) != 3 {
		return KubeaddonsConfigsVersion{},
			fmt.Errorf("version `%s` does not follow standard KubeaddonsConfigsVersion", version)
	}

	distribution := KubeaddonsConfigsDistribution(parts[0])
	if distribution != StableKubeaddonsConfigsDistribution {
		return KubeaddonsConfigsVersion{}, fmt.Errorf("version `%s` does not have a %s distribution", version,
			StableKubeaddonsConfigsDistribution)
	}
	kubernetesVersion, err := semver.Parse(parts[1])
	if err != nil {
		return KubeaddonsConfigsVersion{},
			errors.Wrapf(err, "version `%s`'s kubernetes version part (%s) does not follow semver", version, parts[1])
	}
	releaseNumber, err := strconv.ParseUint(parts[2], 10, 0)
	if err != nil {
		return KubeaddonsConfigsVersion{},
			errors.Wrapf(err, "version `%s`'s kubernetes release number part (%s) is not an unint", version, parts[2])
	}

	return KubeaddonsConfigsVersion{
		Distribution:      distribution,
		KubernetesVersion: kubernetesVersion,
		ReleaseNumber:     releaseNumber,
	}, nil
}
