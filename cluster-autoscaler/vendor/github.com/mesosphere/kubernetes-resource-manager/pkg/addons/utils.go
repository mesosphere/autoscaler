package addons

import (
	"net/url"
	"strings"

	"github.com/mesosphere/kubeaddons/pkg/repositories/git"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
	"github.com/mesosphere/kubeaddons/pkg/events"
	"github.com/mesosphere/kubeaddons/pkg/status"
)

const (
	// ErrorDecodedObjectNotAddonOrClusterAddon is an error message returned
	// if the decoded object is not one we're processing
	ErrorDecodedObjectNotAddonOrClusterAddon = "error: decoded object is " +
		"not a kubeaddons.mesosphere.io Addon or ClusterAddon object"

	// RepositoryURLLabelKey is the label key for the repository url
	RepositoryNameLabelKey = "kubeaddons.mesosphere.io/repository-name"

	// RepositoryRefLabelKey is the label key for the repository ref
	RepositoryRefLabelKey = "kubeaddons.mesosphere.io/repository-ref"
)

// -----------------------------------------------------------------------------
// Utils - Public Functions
// -----------------------------------------------------------------------------

// AddonsAvailable iterates through the TemplateRepos and selects addons that are enabled for the cloud provider.
// it overrides previously existing addons if newer ones are found in higher priority repos.
func AddonsAvailable(provider string, repos TemplateRepos) (map[string]v1beta1.AddonInterface, error) {
	addons := make(map[string]v1beta1.AddonInterface)
	for _, repo := range repos {
		r, err := git.NewRemoteRepository(repo.URL, repo.Tag, "origin")
		if err != nil {
			return nil, errors.Wrapf(err, "the repository %s could not be properly retrieved", repo.URL)
		}

		allAddons, err := r.ListAddons()
		if err != nil {
			return nil, err
		}

		for _, revs := range allAddons {
			latestRev := revs[0]
			if isAddonUsable(provider, latestRev) {
				// append repo labels to addon
				appendRepositoryLabels(latestRev, repo)
				// TODO log a message if an addon already exists in a different repo
				// Need to handle different addon versions coming from the same repo
				// https://jira.mesosphere.com/browse/DCOS-62835
				addons[latestRev.GetName()] = latestRev
			}
		}
	}
	return addons, nil
}

func appendRepositoryLabels(addon v1beta1.AddonInterface, repo TemplateRepo) {
	labels := addon.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// name
	if name, err := urlToName(repo); err == nil {
		labels[RepositoryNameLabelKey] = name
	}

	// ref
	labels[RepositoryRefLabelKey] = tagToLabel(repo.Tag)
	addon.SetLabels(labels)
}

func urlToName(repo TemplateRepo) (string, error) {
	var name string

	url, err := url.Parse(repo.URL)
	if err != nil {
		return name, err
	}

	name = strings.TrimPrefix(url.Path, "/")
	return strings.ReplaceAll(name, "/", "-"), nil
}

func tagToLabel(tag string) string {
	return strings.ReplaceAll(tag, "/", "-")
}

// isAddonUsable determines if the addon is usable based on cloud provider restrictions
func isAddonUsable(provider string, addon v1beta1.AddonInterface) bool {
	usable := false
	providers := addon.GetAddonSpec().CloudProvider
	if len(providers) == 0 {
		usable = true
	} else {
		for _, validProvider := range providers {
			if validProvider.Name == provider {
				usable = true
			}
		}
	}
	return usable
}

// DecodeObjectFromManifest decodes Addons and ClusterAddons from yaml source
func DecodeObjectFromManifest(data []byte) (runtime.Object, error) {
	scheme := runtime.NewScheme()
	if err := v1beta1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// apiextensionsv1beta1.AddToScheme(scheme)
	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if runtime.IsNotRegisteredError(err) {
		return nil, errors.New(ErrorDecodedObjectNotAddonOrClusterAddon)
	}
	if runtime.IsMissingKind(err) {
		return nil, errors.New(ErrorDecodedObjectNotAddonOrClusterAddon)
	}
	return obj, nil
}

func closeAll(cs ...chan events.StatusEvent) {
	for _, c := range cs {
		close(c)
	}
}

func notify(a v1beta1.AddonInterface, s status.Status, cs ...chan events.StatusEvent) {
	e := deployEvent{a, s}
	for _, c := range cs {
		c <- e
	}
}

func listAddonNames(list []v1beta1.AddonInterface) []string {
	addonNames := []string{}
	for _, a := range list {
		addonNames = append(addonNames, a.GetName())
	}
	return addonNames
}
