package git

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	gitutils "github.com/mesosphere/kubeaddons/internal/git/utils"
	internalutils "github.com/mesosphere/kubeaddons/internal/utils"
)

// -----------------------------------------------------------------------------
// Vars & Consts
// -----------------------------------------------------------------------------

var (
	expiry  = time.Minute * 15
	refresh = time.Minute * 3
	wait    = time.Second * 60

	cacheServers = map[string]Cache{}
)

// ErrServerStopped indicates that an operation was attempted on a cache server that was stopped
var ErrServerStopped = errors.New("cache server stopped")

// -----------------------------------------------------------------------------
// Cache
// -----------------------------------------------------------------------------

// Cache represents an in-memory cache of a git repository object tree
type Cache interface {
	Tree() (*object.Tree, error)
	Stop()
}

// NewCache provides a new Cache
func NewCache(giturl, gitref, gitremote string) (Cache, error) {
	k := cacheKey(giturl, gitref, gitremote)
	if c, ok := cacheServers[k]; ok {
		if _, err := c.Tree(); err != nil {
			// if the server is stopped, then we just continue below and recreate
			// otherwise, something has failed and we report it.
			if err != ErrServerStopped {
				return c, err
			}
		} else {
			return c, nil
		}
	}

	repo, err := gitutils.CloneToMemory(giturl)
	if err != nil {
		return nil, err
	}

	ref, err := gitutils.GetRef(repo, gitref)
	if err != nil {
		return nil, err
	}

	purl, err := url.Parse(giturl)
	if err != nil {
		return nil, err
	}

	c := &cache{
		name:   giturl,
		repo:   repo,
		ref:    ref,
		url:    purl,
		remote: gitremote,
		key:    k,

		treeCh: make(chan *object.Tree),
		stopCh: make(chan struct{}),
	}

	if err := c.fetch(); err != nil {
		return c, err
	}

	go c.server()
	return c, nil
}

// -----------------------------------------------------------------------------
// cache - Cache implementation
// -----------------------------------------------------------------------------

func (c *cache) Tree() (*object.Tree, error) {
	if c.stopped {
		return nil, ErrServerStopped
	}

	select {
	case tree := <-c.treeCh:
		return tree, nil
	case _ = <-time.After(wait):
		return nil, fmt.Errorf("timed out (%s) waiting for git tree from cache", wait)
	}
}

func (c *cache) Stop() {
	c.stopCh <- struct{}{}
	close(c.stopCh)
	c.stopped = true
}

// -----------------------------------------------------------------------------
// cache - Private
// -----------------------------------------------------------------------------

type cache struct {
	name   string
	repo   *git.Repository
	tree   *object.Tree
	ref    *plumbing.Reference
	url    *url.URL
	remote string
	key    string

	treeCh  chan *object.Tree
	stopCh  chan struct{}
	stopped bool

	expiry *time.Time
}

func (c *cache) fetch() error {
	if err := c.repo.Fetch(&git.FetchOptions{RemoteName: c.remote}); err != nil {
		if err.Error() != git.NoErrAlreadyUpToDate.Error() {
			return err
		}
	}

	tree, err := gitutils.GetRefTree(c.repo, c.ref)
	if err != nil {
		return err
	}
	c.tree = tree

	return nil
}

func (c *cache) server() {
	for {
		select {
		case c.treeCh <- c.tree:
			newCleanupTime := time.Now().Add(expiry)
			c.expiry = &newCleanupTime
			continue
		case <-time.After(refresh):
			if time.Now().After(*c.expiry) {
				fmt.Printf("cache for (%s) has expired", c.key)
				return
			}
			c.fetch()
		case <-c.stopCh:
			return
		}
	}
}

func cacheKey(url, ref, remote string) string {
	return fmt.Sprintf("url: %s ref: %s remote: %s", url, ref, remote)
}

func init() {
	overrides := map[string]*time.Duration{
		"KUBEADDONS_GIT_CACHE_SERVER_EXPIRY":             &expiry,
		"KUBEADDONS_GIT_CACHE_SERVER_REFRESH":            &refresh,
		"KUBEADDONS_GIT_CACHE_SERVER_RESPONSE_WAIT_TIME": &wait,
	}
	for env, duration := range overrides {
		if err := internalutils.EnvDuration(env, duration); err != nil {
			panic(err)
		}
	}
}
