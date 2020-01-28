package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/blang/semver"
	"github.com/mesosphere/kubeaddons/internal/k8s"
	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/constants"
	"github.com/mesosphere/kubeaddons/pkg/status"
)

// -----------------------------------------------------------------------------
// Harness - Types
// -----------------------------------------------------------------------------

// Harness is a test harness for managing and automating addon tests.
// It comes loaded with some pre-defined generic addon tests.
//
// TODO: later we want to enable a test.Plan type which the Harness can load, and make the existing
// plans (validate, deploy, cleanup) default modules. See (https://jira.mesosphere.com/browse/DCOS-61266)
type Harness interface {
	// Validate runs the addons provided to the test.Harness through various validation checks
	Validate()

	// Deploy tests the deployment of all addons provided to the test.Harness, ensuring that
	// the addons become active, and healthy.
	Deploy()

	// Cleanup tests the teardown of all addons provided to the test.Harness
	Cleanup()
}

// -----------------------------------------------------------------------------
// Basic Harness - Public
// -----------------------------------------------------------------------------

// NewBasicTestHarness is an implementation of test.Harness which focuses on simplicity
func NewBasicTestHarness(t *testing.T, cluster Cluster, addons ...v1beta1.AddonInterface) (Harness, error) {
	// ensure at least one addon is present
	if len(addons) == 0 {
		return nil, fmt.Errorf("no addons provided to test harness")
	}

	// ensure there are no duplicates, and store provided addons in a map
	maddons := map[string]v1beta1.AddonInterface{}
	for _, addon := range addons {
		if _, ok := maddons[addon.GetName()]; ok {
			return nil, fmt.Errorf("found duplicate of addon %s", addon.GetName())
		}
		maddons[addon.GetName()] = addon
	}

	// get the dynamic k8s client needed for addons
	c, err := k8s.DynamicClient(cluster.Config())
	if err != nil {
		return nil, err
	}

	return &basicTestHarness{
		addons:  maddons,
		cluster: cluster,
		t:       t,
		c:       c,
	}, nil
}

func (th *basicTestHarness) Validate() {
	defer func() {
		if r := recover(); r != nil {
			th.t.Fatalf("had to recover in validate test plan: %s", r)
		}
	}()

	// ensure addon-revision is set
	for name, addon := range th.addons {
		if revision, ok := addon.GetAnnotations()[constants.AddonRevisionAnnotation]; ok {
			if _, err := semver.Parse(strings.TrimPrefix(revision, "v")); err != nil {
				th.t.Fatalf("addon %s had an invalid addon-revision annotation: %s", name, revision)
			}
		} else {
			th.t.Fatalf("addon %s does not have addon-revision information set", name)
		}
	}

	// TODO - dependency validation, see: https://jira.mesosphere.com/browse/DCOS-61287
	// TODO - airgap capabilities validation, see: https://jira.mesosphere.com/browse/DCOS-61218
}

func (th *basicTestHarness) Deploy() {
	defer func() {
		if r := recover(); r != nil {
			th.t.Fatalf("had to recover in deploy test plan: %s", r)
		}
	}()

	for name, addon := range th.addons {
		if err := th.c.Create(context.TODO(), addon); err != nil {
			th.t.Errorf("could not create addon %s: %w", name, err)
		}
	}

	addonscpy := map[string]v1beta1.AddonInterface{}
	for k, v := range th.addons {
		addonscpy[k] = v
	}

	nextLog := time.Now()
	timeout := time.Now().Add(defaultSetupWaitDuration + (defaultAddonWaitDuration * time.Duration(len(addonscpy))))
	for timeout.After(time.Now()) {
		if len(addonscpy) == 0 {
			break
		}
		if time.Now().After(nextLog) {
			th.t.Logf("STATUS: deploying addons: %+v", addonscpy)
			nextLog = time.Now().Add(time.Minute * 2)
		}
		for name, addon := range addonscpy {
			updated := addon.NewEmptyType()
			if err := th.c.Get(context.TODO(), types.NamespacedName{
				Name:      name,
				Namespace: addon.GetNamespace(),
			}, updated); err != nil {
				th.t.Errorf("could not retrieve addon %s (namespace: %s): %w", name, addon.GetNamespace(), err)
			}

			if updated.GetStatus().Ready && updated.GetStatus().Stage == status.Deployed {
				delete(addonscpy, name)
			}
		}
	}

	if len(addonscpy) > 0 {
		for k, v := range addonscpy {
			th.t.Errorf("addon %s did not deploy properly. Status: %+v", k, v.GetStatus())
		}
	} else {
		th.t.Logf("STATUS: addons deployed successfully")
	}
}

func (th *basicTestHarness) Cleanup() {
	defer func() {
		if r := recover(); r != nil {
			th.t.Fatalf("had to recover in cleanup test plan: %s", r)
		}
	}()

	for name, addon := range th.addons {
		if err := th.c.Delete(context.TODO(), addon); err != nil {
			if !errors.IsNotFound(err) {
				th.t.Errorf("could not delete addon %s: %w", name, err)
			}
		}
	}

	addonscpy := map[string]v1beta1.AddonInterface{}
	for k, v := range th.addons {
		addonscpy[k] = v
	}

	nextLog := time.Now()
	timeout := time.Now().Add(defaultSetupWaitDuration + (defaultAddonWaitDuration * time.Duration(len(addonscpy))))
	for timeout.After(time.Now()) {
		if len(addonscpy) == 0 {
			break
		}
		if time.Now().After(nextLog) {
			th.t.Logf("STATUS: cleaning up addons: %+v", addonscpy)
			nextLog = time.Now().Add(time.Minute * 2)
		}
		for name, addon := range addonscpy {
			updated := addon.NewEmptyType()
			if err := th.c.Get(context.TODO(), types.NamespacedName{
				Name:      name,
				Namespace: addon.GetNamespace(),
			}, updated); err != nil {
				if errors.IsNotFound(err) {
					delete(addonscpy, name)
				} else {
					th.t.Errorf("could not retrieve addon %s: %w", name, err)
				}
			}
		}
	}

	if len(addonscpy) > 0 {
		for k, v := range addonscpy {
			th.t.Errorf("addon %s did not delete properly. Status: %+v", k, v.GetStatus())
		}
	} else {
		th.t.Logf("STATUS: addons cleaned up successfully")
	}
}

// -----------------------------------------------------------------------------
// Basic Harness - Private
// -----------------------------------------------------------------------------

type basicTestHarness struct {
	addons  map[string]v1beta1.AddonInterface
	cluster Cluster

	t *testing.T
	c client.Client
}
