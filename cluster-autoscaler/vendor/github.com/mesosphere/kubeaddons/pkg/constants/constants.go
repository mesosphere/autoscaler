package constants

import (
	"time"
)

const (
	// DefaultAddonNamespace indicates the namespace that should be used for addons if none is defined
	DefaultAddonNamespace = "kubeaddons"

	// DefaultAddonReleaseSuffix is a prepending string added to addons deployed via helm
	DefaultAddonReleaseSuffix = "kubeaddons"

	// DefaultConfigRepo is the default addon configuration repo to use
	DefaultConfigRepo = "https://github.com/mesosphere/kubernetes-base-addons"

	// DefaultDeploymentWaitTimePerAddon is the maximum amount of time to wait for the deployment of any addon by default
	DefaultDeploymentWaitTimePerAddon = time.Minute * 5
)
