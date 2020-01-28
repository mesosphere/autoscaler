package events

import (
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/status"
)

// StatusEvent is an interface which describes an update to an addons status
type StatusEvent interface {
	// Addon provides the addon that this event is for
	Addon() v1beta1.AddonInterface

	// Status provides the status of the Addon when this Event fired
	Status() status.Status
}
