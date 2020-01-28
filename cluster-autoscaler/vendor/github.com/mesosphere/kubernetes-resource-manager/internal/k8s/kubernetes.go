package k8s

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/mesosphere/kubeaddons/pkg/api/v1beta1"
)

// -----------------------------------------------------------------------------
// Utils - Kubernetes Client
// -----------------------------------------------------------------------------

// KubeConfigPath returns the path of the kubeconfig
func KubeConfigPath() (string, error) {
	log := crlog.Log.WithName("KubeConfigPath")
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		h, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		kubeconfig = filepath.Join(h, ".kube", "config")
	}
	_, err := os.Stat(kubeconfig)
	switch {
	case os.IsNotExist(err):
		// The config file does not exist, return an empty string so we can configure for in-cluster
		log.Info("configuring client for in-cluster")
		return "", nil
	case err != nil:
		log.Info("stat failed reading config file", "error", err, "config path", kubeconfig)
		return "", err
	default:
		log.Info("configuring client for out-of-cluster", "config path", kubeconfig)
	}
	return kubeconfig, nil
}

// DefaultRestConfig provides a *rest.Config
func DefaultRestConfig() (*rest.Config, error) {
	kubeconfig, err := KubeConfigPath()
	if err != nil {
		return nil, err
	}
	if kubeconfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// DefaultClient provides a kubernetes.Interface using the DefaultRestConfig
func DefaultClient(conf *rest.Config) (kubernetes.Interface, error) {
	if conf == nil {
		var err error
		conf, err = DefaultRestConfig()
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(conf)
}

// DynamicClient provides a client.Client using the DefaultRestConfig
func DynamicClient(conf *rest.Config) (client.Client, error) {
	if conf == nil {
		var err error
		conf, err = DefaultRestConfig()
		if err != nil {
			return nil, err
		}
	}
	scheme := runtime.NewScheme()
	if err := kscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := v1beta1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(conf, client.Options{
		Scheme: scheme,
	})
}
