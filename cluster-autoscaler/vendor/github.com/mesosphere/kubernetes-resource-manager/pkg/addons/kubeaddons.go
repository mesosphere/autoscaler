package addons

import (
	"context"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kerrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/chartutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/constants"
	addonerrs "github.com/mesosphere/kubeaddons/pkg/errors"
	"github.com/mesosphere/kubeaddons/pkg/events"
	"github.com/mesosphere/kubeaddons/pkg/status"

	"github.com/mesosphere/kubernetes-resource-manager/internal/k8s"
)

// New provides a list of v1beta1.Addons given a list of AddonConfigs and
// TemplateRepos, using the TemplateRepos to find all available addons in
// all repos, and the AddonConfigs to disable or reconfigure those addons
// where needed.
func New(provider string, cfgs AddonConfigs, repos TemplateRepos) (*Kubeaddons, error) {
	addonsAvailable, err := AddonsAvailable(provider, repos)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available Addons: %+v", err)
	}

	addons := make([]v1beta1.AddonInterface, 0, len(cfgs))
	for _, cfg := range cfgs {
		// pass if this addon is not even enabled
		if !cfg.Enabled {
			continue
		}

		// ensure that the addon is available in the repos or throw an error
		addon, ok := addonsAvailable[cfg.Name]
		if !ok {
			continue
		}

		if cfg.Values != "" {
			if addon.GetAddonSpec().ChartReference.Values != nil {
				values, err := chartutil.ReadValues([]byte(*addon.GetAddonSpec().ChartReference.Values))
				if err != nil {
					return nil, fmt.Errorf("error decoding values from Addon template: %v", err)
				}
				configValues, err := chartutil.ReadValues([]byte(cfg.Values))
				if err != nil {
					return nil, fmt.Errorf("error decoding values from config file: %v", err)
				}
				values.MergeInto(configValues)
				mergedValues, err := values.YAML()
				if err != nil {
					return nil, fmt.Errorf("error merging configured values with template values: %v", err)
				}
				addon.GetAddonSpec().ChartReference.Values = &mergedValues
			} else {
				values := cfg.Values
				addon.GetAddonSpec().ChartReference.Values = &values
			}
		}

		// add this addon to the list of addons for deployment
		addons = append(addons, addon)
	}

	cfg, err := k8s.DefaultRestConfig()
	if err != nil {
		return &Kubeaddons{Addons: addons}, err
	}

	_, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return &Kubeaddons{Addons: addons, RestConfig: cfg}, err
	}

	return &Kubeaddons{Addons: addons, RestConfig: cfg}, nil
}

// Cleanup will work through an AddonList removing any "disabled" addons from the cluster
func (d Kubeaddons) Cleanup(provider string, cfgs AddonConfigs, repos TemplateRepos) error {
	addonsAvailable, err := AddonsAvailable(provider, repos)
	if err != nil {
		return fmt.Errorf("failed to fetch available Addons: %+v", err)
	}

	dynamicClient, err := k8s.DynamicClient(d.RestConfig)
	if err != nil {
		return err
	}

	for _, cfg := range cfgs {
		if !cfg.Enabled {
			addon, ok := addonsAvailable[cfg.Name]
			if !ok {
				continue
			}
			currentAddon := addon.NewEmptyType()

			// in our existing kubeaddons-configs, we use the "kubeaddons"
			// namespace, we may wish to move it to cluster scoped in the
			// future
			objKey := types.NamespacedName{
				Name:      addon.GetName(),
				Namespace: addon.GetNamespace(),
			}
			err := dynamicClient.Get(context.TODO(), objKey, currentAddon)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return err
			}
			err = dynamicClient.Delete(context.TODO(), currentAddon)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Apply synchronously deploys all configured addons
// TODO: it should also remove installed Addons that are no longer desired
func (d Kubeaddons) Apply(out io.Writer, verbose bool, repoOverride string) error {
	dynamicClient, err := k8s.DynamicClient(d.RestConfig)
	if err != nil {
		return err
	}

	for _, addon := range d.Addons {
		// used for konvoy airgapped, if a repo override has been provided
		// use that in place of any chart repository.
		if repoOverride != "" {
			if chartRef := addon.GetAddonSpec().ChartReference; chartRef != nil {
				stripped := strings.TrimPrefix(chartRef.Chart, "stable/")
				stripped = strings.TrimPrefix(stripped, "staging/")
				chartRef.Repo = &repoOverride
				chartRef.Chart = stripped
			}
		}

		err := dynamicClient.Create(context.Background(), addon)
		if err != nil {
			// if the error IsAlreadyExists we want to try and update the addon
			// with any new values from this deployment instead. Otherwise its
			// still an error condition and we bail.
			if kerrs.IsAlreadyExists(err) {
				// Retry updates if "the object has been modified"
				for retries := 0; retries < 10; retries++ {
					currentAddon := addon.NewEmptyType()
					objKey := types.NamespacedName{
						Name:      addon.GetName(),
						Namespace: addon.GetNamespace(),
					}
					err := dynamicClient.Get(context.TODO(), objKey, currentAddon)
					if err != nil {
						if verbose {
							fmt.Fprintf(out, "\nUnable to get the state of the addon: %s with error: %v", addon.GetName(), err)
						}
					}
					addon.SetResourceVersion(currentAddon.GetResourceVersion())

					err = dynamicClient.Update(context.TODO(), addon)
					if err != nil && strings.Contains(err.Error(), registry.OptimisticLockErrorMsg) {
						continue
					} else if err != nil {
						return err
					}

					// successful update, let's reset the status now to the new addon
					if !reflect.DeepEqual(addon.GetAddonSpec(), currentAddon.GetAddonSpec()) {
						if verbose {
							fmt.Fprintf(out, "\nAddon is different reseting its status to false with name: %s\n", addon.GetName())
						}
						// Get latest changes to be able to update the status
						objKey := types.NamespacedName{
							Name:      addon.GetName(),
							Namespace: addon.GetNamespace(),
						}
						err := dynamicClient.Get(context.TODO(), objKey, addon)
						if err != nil {
							if verbose {
								fmt.Fprintf(out, "\nUnable to get the state of the addon: %s with error: %v", addon.GetName(), err)
							}
						}
						addon.GetStatus().Ready = false
						addon.GetStatus().Stage = status.Empty
						err = dynamicClient.Status().Update(context.TODO(), addon)
						if err != nil {
							if verbose {
								fmt.Fprintf(out, "\nFailed to update addon status: %s with error: %v", addon.GetName(), err)
							}
							// failed to update the status, retry again
							continue
						}
					}
					break
				}
			} else {
				if verbose {
					fmt.Fprintf(out, "\nFailed to create addon status: %s with error: %v", addon.GetName(), err)
				}
				return err
			}
		}
	}
	return nil
}

// Wait waits for the addons to be deployed and reports errors for any that do not deploy properly
func (d Kubeaddons) Wait(out io.Writer, verbose bool, cs ...chan events.StatusEvent) error {
	defer closeAll(cs...)
	dynamicClient, err := k8s.DynamicClient(d.RestConfig)
	if err != nil {
		return err
	}
	// Wait for the controller to be running prior start waiting for the addon
	// to be deployed.
	if err := d.waitForControllerToBeRunning(dynamicClient, out, verbose); err != nil {
		return err
	}

	remainingAddons := make([]v1beta1.AddonInterface, len(d.Addons))
	copy(remainingAddons, d.Addons)
	// Using a logarithmic multiplier will more closely model the advantages
	// of parallel installations over that of a linear multiple. This will
	// produce a timeout range from 1x for one addon, to 3.7x for 15 addons.
	timeout := time.Now().Add(constants.DefaultDeploymentWaitTimePerAddon *
		time.Duration(math.Log(float64(len(remainingAddons)))+1))
	for range d.Addons {
		timeout.Add(constants.DefaultDeploymentWaitTimePerAddon)
	}

	lastNotify := map[string]status.Status{}
	var errs *addonerrs.Stack
	finished := false
	if verbose {
		fmt.Fprintf(out, "Waiting for the addons to be deployed until: %v\n", timeout)
	}
	for timeout.After(time.Now()) && len(remainingAddons) != 0 {
		var newRemainingAddons []v1beta1.AddonInterface
		for i, addon := range remainingAddons {
			r := addon.NewEmptyType()
			objKey := types.NamespacedName{
				Name:      addon.GetName(),
				Namespace: addon.GetNamespace(),
			}
			err := dynamicClient.Get(context.Background(), objKey, r)
			if err != nil {
				return fmt.Errorf("fatal: the addon, %s, has been removed unexpectedly: %v", addon.GetName(), err)
			}
			if r.GetStatus().Stage != lastNotify[addon.GetName()] {
				lastNotify[addon.GetName()] = r.GetStatus().Stage
				if verbose {
					fmt.Fprintf(out, "\nStatus change: %v; with name: %s\n", r.GetStatus().Stage, r.GetName())
				}
			}
			switch r.GetStatus().Stage {
			case status.Deployed:
				notify(addon, r.GetStatus().Stage, cs...)
				continue
			case status.Failed:
				if verbose {
					fmt.Fprintf(out, "\nStatus initially failed %s, checking the state of the dependencies to confirm\n", r.GetName())
				}

				if len(addon.GetAddonSpec().Requires) != 0 {
					ready, err := d.areAddonDependenciesReady(out, verbose, dynamicClient, addon)
					if ready {
						if verbose {
							fmt.Fprintf(out, "\nAddon dependencies are ready for addon name: %s\n", r.GetName())
						}
						// NOTE: Dependencies are ready but we are gonna wait a bit more
						err := d.waitForAddon(out, verbose, dynamicClient, addon, KubeaddonsControllerRunningTimeout)
						if err != nil {
							// deployed failed or timeout, let's report a failure issue
							addonerrs.Push(&errs, fmt.Errorf("addon %s has failed", addon.GetName()))
						}
						notify(addon, r.GetStatus().Stage, cs...)
						continue
					}

					if verbose && (err != nil || !ready) {
						fmt.Fprintf(out, "\nAddon dependencies are not ready or we failed to fetch its dependencies: %s\n", r.GetName())
					}
				} else {
					// NOTE: If it doesn't have dependencies then wait a bit to determine a failed deployment
					err := d.waitForAddon(out, verbose, dynamicClient, addon, KubeaddonsControllerRunningTimeout)
					// If error is not nil, then the function timeout
					if err != nil {
						fmt.Fprintf(out, "\nAddon failed the deployment with name: %s\n", r.GetName())
						notify(addon, r.GetStatus().Stage, cs...)
						// deployed failed or timeout, let's report a failure issue
						addonerrs.Push(&errs, fmt.Errorf("addon %s has failed", addon.GetName()))
						continue
					}
				}
			}
			newRemainingAddons = append(newRemainingAddons, remainingAddons[i])
		}
		remainingAddons = newRemainingAddons
		if verbose {
			fmt.Fprintf(out, "\nWaiting for the following remaining addons to be deployed or failed: %v\n",
				listAddonNames(remainingAddons))
		}
	}

	finished = len(remainingAddons) == 0
	if !finished {
		var names []string
		for _, a := range remainingAddons {
			names = append(names, a.GetName())
		}
		addonerrs.Push(&errs, fmt.Errorf("deployment timed out! the following addons were not resolved: %s",
			strings.Join(names, ", ")))
	}

	return errs
}

// waitForAddon waits until a addon is finally deployed, otherwise it'll timeout
// NOTE: we need this until ready state of an addon means the app is ready and not just deployed.
func (d Kubeaddons) waitForAddon(
	out io.Writer, verbose bool, dynamicClient client.Client, addon v1beta1.AddonInterface, timeout time.Duration) error {
	if verbose {
		fmt.Fprintf(out, "\nWaiting for addon '%s' to fail or be deployed\n", addon.GetName())
	}
	return wait.Poll(TimeToWaitForAddonInterval, timeout, func() (bool, error) {
		r := addon.NewEmptyType()
		objKey := types.NamespacedName{
			Name:      addon.GetName(),
			Namespace: addon.GetNamespace(),
		}
		err := dynamicClient.Get(context.Background(), objKey, r)
		if err != nil {
			return false, nil
		}
		if verbose {
			fmt.Fprintf(out, "\nRetry waiting for addon '%s' to fail or be deployed with status: %v...\n", r.GetName(),
				r.GetStatus().Stage)
		}
		switch r.GetStatus().Stage {
		case status.Deployed:
			return true, nil
		case status.Failed:
			return false, nil
		}
		return false, nil
	})
}

// areAddonDependenciesReady checks if the dependencies are deployed
func (d Kubeaddons) areAddonDependenciesReady(
	out io.Writer, verbose bool, dynamicClient client.Client, addon v1beta1.AddonInterface) (bool, error) {
	for _, require := range addon.GetAddonSpec().Requires {
		deps, err := v1beta1.GetDeps(dynamicClient, &require)
		if err != nil {
			return false, err
		}

		numberReady := 0
		for _, dep := range deps {
			if dep.GetStatus().Ready {
				numberReady += 1
				if verbose {
					fmt.Fprintf(out, "1 of %v dependencies ready, dependencies are satisfied: %s", len(deps), dep.GetName())
				}
			} else if verbose {
				fmt.Fprintf(out, "1 of %v dependencies not ready: %s", len(deps), dep.GetName())
			}
		}
		if numberReady == 0 {
			return false, nil
		}
	}
	return true, nil
}

// waitForControllerToBeRunning waits for the controller to be running
func (d Kubeaddons) waitForControllerToBeRunning(dynamicClient client.Client, out io.Writer, verbose bool) error {
	lastKnownPodNumber := -1
	label := make(map[string]string)
	label["control-plane"] = KubeaddonsControllerLabel

	err := wait.PollImmediate(APICallRetryInterval, KubeaddonsControllerRunningTimeout, func() (bool, error) {
		pods := &corev1.PodList{}
		err := dynamicClient.List(context.Background(), pods, client.MatchingLabels(label))
		if err != nil {
			if verbose {
				fmt.Fprintf(out, "Error getting controller pod with label selector %q [%v]\n", KubeaddonsControllerLabel, err)
			}
			return false, nil
		}

		if lastKnownPodNumber != len(pods.Items) {
			if verbose {
				fmt.Fprintf(out, "Found controller pod for label selector %s\n", KubeaddonsControllerLabel)
			}
			lastKnownPodNumber = len(pods.Items)
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("kubeaddons controller is not running yet")
	}
	return nil
}
