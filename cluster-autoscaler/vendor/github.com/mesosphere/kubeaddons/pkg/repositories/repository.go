package repositories

import (
	"github.com/mesosphere/kubeaddons/pkg/repositories/addons/revisions"
)

// RepositoryList is a collection of several Repository resources
type RepositoryList map[string]Repository

// Repository represents a store of Addon resources, metadata, and other files
type Repository interface {
	// Name provides the user-readable identifier for the Repository
	Name() string

	// GetAddon retrieves a specific addon and all of its revisions by name
	GetAddon(string) (revisions.AddonRevisions, error)

	// ListAddons lists all addons present in the repository
	ListAddons(filters ...AddonFilter) (map[string]revisions.AddonRevisions, error)
}
