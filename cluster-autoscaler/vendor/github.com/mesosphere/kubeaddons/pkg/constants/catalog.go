package constants

import "regexp"

const (
	// AddonOriginRepositoryAnnotation indicates the annotation that describes the repository the addon originates from
	AddonOriginRepositoryAnnotation = "catalog.kubeaddons.mesosphere.io/origin-repository"

	// AddonOriginRepositoryVersionAnnotation indicates the annotation that describes the repository version the addon originates from
	AddonOriginRepositoryVersionAnnotation = "catalog.kubeaddons.mesosphere.io/origin-repository-version"

	// AddonRevisionAnnotation
	AddonRevisionAnnotation = "catalog.kubeaddons.mesosphere.io/addon-revision"

	// DefaultRepositoryVersionPrefix is the prefex that comes with a Ref name of an Repository
	DefaultRepositoryVersionPrefix = `stable-`

	// RepositoryConfigurationExitCode is the exit code that indicates the API could not be started due to a problem with configured repositories
	RepositoryConfigurationExitCode = 111

	// KubeAddonsConfigsRepository defines a default URL where Addons can be pulled from
	KubeAddonsConfigsRepository = "https://github.com/mesosphere/kubernetes-base-addons"

	//KubeAddonsCommunityRepository defines the URL for the community repo used for testing
	KubeAddonsCommunityRepository = "https://github.com/mesosphere/kubeaddons-community"

	// DemoAddonRepository is the URL of the repo for demo purposes
	DemoAddonRepository = "https://github.com/mesosphere/kubeaddons-demo"

	// DefaultTestingAddonRepositoryRef defines a default version for the KubeAddonsConfigsRepository
	DefaultTestingAddonRepositoryRef = "master"

	// DemoAddonRepositoryRef defines the repo version for demo-ing
	DemoAddonRepositoryRef = DefaultTestingAddonRepositoryRef

	// EnvironmentTesting is the testing environment
	EnvironmentTesting = "testing"

	// EnvironmentDevelopment is the development environment
	EnvironmentDevelopment = "development"
)

var (
	// StableRepositoryVersionRegex indicates the format the name of the ref should be in to indicate it's a stable release of the underlying software it matches either tags or remote branches
	StableRepositoryVersionRegex = regexp.MustCompile(`^refs\/(tags|remotes\/origin)\/stable-[0-9]+\.[0-9]+(\.[0-9]+)?`)
)
