package addons

import (
	"time"

	"k8s.io/client-go/rest"

	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/status"
)

const (
	// TimeToWaitForAddonFailures time to determine an addon failed its deployment
	TimeToWaitForAddonFailures = 3 * time.Minute
	// TimeToWaitForAddonInterval time to wait to check if an addon is deployed
	TimeToWaitForAddonInterval = 20 * time.Second
	// APICallRetryInterval defines how long kubeaddons should wait before retrying a failed API operation
	APICallRetryInterval = 500 * time.Millisecond
	// KubeaddonsControllerLabel label of the controller
	KubeaddonsControllerLabel = "kubeaddons-controller-manager"
	// KubeaddonsControllerRunningTimeout time to wait until the controller is running
	KubeaddonsControllerRunningTimeout = 3 * time.Minute
)

// -----------------------------------------------------------------------------
// Kubeaddons - Public Interface
// -----------------------------------------------------------------------------

// Kubeaddons represents a range of addons that can be deployed to a Kubernetes cluster
type Kubeaddons struct {
	Addons []v1beta1.AddonInterface

	// RestConfig will be defaulted in New. Override this after creating a new
	// Kubeaddons resource if you need a custom rest.Config
	RestConfig *rest.Config
}

type deployEvent struct {
	addon  v1beta1.AddonInterface
	status status.Status
}

func (d deployEvent) Addon() v1beta1.AddonInterface {
	return d.addon
}

func (d deployEvent) Status() status.Status {
	return d.status
}
