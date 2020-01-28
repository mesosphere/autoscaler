package git

import (
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	gitutils "github.com/mesosphere/kubeaddons/internal/git"
	internalutils "github.com/mesosphere/kubeaddons/internal/utils"
	"github.com/mesosphere/kubeaddons/pkg/errors"
	"github.com/mesosphere/kubeaddons/pkg/repositories"
	"github.com/mesosphere/kubeaddons/pkg/repositories/addons/revisions"
)

// -----------------------------------------------------------------------------
// Git - Public
// -----------------------------------------------------------------------------

// NewRemoteRepository provides a new Repository backed by a remote Git
func NewRemoteRepository(url, ref, remote string) (repositories.Repository, error) {
	cache, err := gitutils.NewCache(url, ref, remote)
	if err != nil {
		return nil, err
	}
	return &git{url: url, ref: ref, remote: remote, cache: cache}, nil
}

// -----------------------------------------------------------------------------
// Git - Public - Repository Implementation
// -----------------------------------------------------------------------------

func (g *git) Name() string {
	return g.url
}

func (g *git) GetAddon(name string) (revisions.AddonRevisions, error) {
	addons, err := g.ListAddons()
	if err != nil {
		return nil, err
	}

	revisions, ok := addons[name]
	if !ok {
		return revisions, errors.ErrorAddonNotFound
	}

	return revisions, nil
}

func (g *git) ListAddons(filters ...repositories.AddonFilter) (map[string]revisions.AddonRevisions, error) {
	addons := map[string]revisions.AddonRevisions{}
	tree, err := g.tree()
	if err != nil {
		return nil, err
	}

	err = tree.Files().ForEach(func(o *object.File) error {
		contents, err := o.Contents()
		if err != nil {
			return err
		}

		isAddon, addon, err := internalutils.IsThisAnAddon([]byte(contents))
		if err != nil {
			if err.Error() == errors.ErrorDecodedObjectNotAddonOrClusterAddon {
				return nil
			}
			return err
		}

		if !isAddon {
			return nil
		}

		if repositories.IncludeAddon(addon, filters) {
			addons[addon.GetName()] = append(addons[addon.GetName()], addon)
		}

		return nil
	})
	if err != nil {
		return addons, err
	}

	for _, revisions := range addons {
		if err := internalutils.SortAddonRevisions(revisions); err != nil {
			return addons, err
		}
	}

	return addons, nil
}

// -----------------------------------------------------------------------------
// Git - Private
// -----------------------------------------------------------------------------

type git struct {
	url    string
	ref    string
	remote string
	cache  gitutils.Cache
}

func (g *git) tree() (*object.Tree, error) {
	tree, err := g.cache.Tree()
	if err != nil {
		if err == gitutils.ErrServerStopped {
			g.cache, err = gitutils.NewCache(g.url, g.ref, g.remote)
			if err != nil {
				return nil, err
			}
			return g.cache.Tree()
		}
	}
	return tree, nil
}
