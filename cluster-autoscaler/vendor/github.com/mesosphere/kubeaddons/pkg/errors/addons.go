package errors

import (
	"errors"
)

const (
	// ErrorDecodedObjectNotAddonOrClusterAddon is an error message returned if the decoded object is not one we're processing
	ErrorDecodedObjectNotAddonOrClusterAddon = "error: decoded object is not a kubeaddons.mesosphere.io Addon or ClusterAddon object"
)

var (
	// ErrorAddonNotFound is an error describing when an addon requested by an end user was not found in the provided repository
	ErrorAddonNotFound = errors.New("addon not found")
)
