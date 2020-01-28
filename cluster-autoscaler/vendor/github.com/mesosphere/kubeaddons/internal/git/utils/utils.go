package utils

import (
	"fmt"
	"path"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// CloneToMemory is an opinioned repo cloner which clones the repo to memory
func CloneToMemory(repoURL string) (*git.Repository, error) {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: repoURL})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GetRef provides a *plumbing.Reference given a repository and refname
func GetRef(r *git.Repository, refname string) (*plumbing.Reference, error) {
	ref, err := r.Tag(refname)
	if err != nil {
		if err == git.ErrTagNotFound {
			ref, err = r.Reference(plumbing.ReferenceName(path.Join("refs", "remotes", "origin", refname)), false)
			if err != nil {
				return nil, fmt.Errorf("%s: ref(%s)", err, refname)
			}
		} else {
			return nil, fmt.Errorf("%s: ref(%s)", err, refname)
		}
	}
	return ref, nil
}

// GetRefTree provides a git *object.Tree for the provided *git.Repository given a *plumbing.Reference
func GetRefTree(r *git.Repository, ref *plumbing.Reference) (*object.Tree, error) {
	// retrieve the commit object for the ref
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		if err != plumbing.ErrObjectNotFound {
			return nil, fmt.Errorf("error with commit object %s: %w", ref.Hash(), err)
		}

		// annotated tag
		tag, err := r.TagObject(ref.Hash())
		if err != nil {
			return nil, fmt.Errorf("error finding tag object after finding commit object %s: %w", ref.Hash(), err)
		}
		commit, err = r.CommitObject(tag.Target)
		if err != nil {
			return nil, fmt.Errorf("error with commit object of a tag %s: %w", ref.Hash(), err)
		}
	}

	// get the commit tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// ListRefs provides a list of refs for a repository, optionally matching the name against any
// number of provided *regexp.Regexp via MatchString().
func ListRefs(r *git.Repository, rgxs ...*regexp.Regexp) ([]*plumbing.Reference, error) {
	refs, err := r.References()
	if err != nil {
		return nil, err
	}
	defer refs.Close()

	filteredRefs := []*plumbing.Reference{}
	refs.ForEach(func(ref *plumbing.Reference) error {
		for _, r := range rgxs {
			if !r.MatchString(ref.Name().String()) {
				return nil
			}
		}
		filteredRefs = append(filteredRefs, ref)
		return nil
	})

	return filteredRefs, nil
}
