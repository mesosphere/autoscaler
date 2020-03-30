package konvoy

import (
	"fmt"

	kommanderv1beta1 "github.com/mesosphere/kommander-cluster-lifecycle/clientapis/pkg/apis/kommander/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/klog"
)

const (
	// ProviderName is the cloud provider name for konvoy
	ProviderName = "konvoy"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "konvoy.d2iq.com/gpu"

	// KubernetesMasterNodeLabel marks nodes that runs master services.
	KubernetesMasterNodeLabel = "node-role.kubernetes.io/master"

	// KonvoyNodeAnnotationKey is annotation that is set on kubernetes nodes
	// that needs to be delete.
	KonvoyNodeAnnotationKey = "konvoy.d2iq.io/delete-machine"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
	}
)

// KonvoyCloudProvider implements CloudProvider interface for konvoy
type KonvoyCloudProvider struct {
	konvoyManager   *KonvoyManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildKonvoyCloudProvider builds a CloudProvider for konvoy. Builds
// node groups from passed in specs.
func BuildKonvoyCloudProvider(konvoyManager *KonvoyManager, do cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) (*KonvoyCloudProvider, error) {
	konvoy := &KonvoyCloudProvider{
		konvoyManager:   konvoyManager,
		resourceLimiter: resourceLimiter,
	}

	if len(do.NodeGroupSpecs) > 0 {
		return nil, fmt.Errorf("KonvoyCloudProvider does not support static node groups")
	}

	return konvoy, nil
}

// Name returns name of the cloud provider.
func (konvoy *KonvoyCloudProvider) Name() string {
	return ProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (konvoy *KonvoyCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (konvoy *KonvoyCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// NodeGroups returns all node groups configured for this cloud provider.
func (konvoy *KonvoyCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	ngs := konvoy.konvoyManager.GetNodeGroups()
	out := make([]cloudprovider.NodeGroup, len(ngs))
	copy(out, ngs)
	return out
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (konvoy *KonvoyCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// NodeGroupForNode returns the node group for the given node.
func (konvoy *KonvoyCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	if _, found := node.ObjectMeta.Labels[KubernetesMasterNodeLabel]; found {
		return nil, nil
	}

	nodeName := kubernetesNodeName(node, konvoy.konvoyManager.provisioner)
	nodeGroupName, err := konvoy.konvoyManager.GetNodeGroupNameForNode(nodeName)
	if err != nil {
		return nil, err
	}

	for _, nodeGroup := range konvoy.konvoyManager.GetNodeGroups() {
		if nodeGroup.Name == nodeGroupName {
			return nodeGroup, nil
		}
	}
	return nil, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (konvoy *KonvoyCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (konvoy *KonvoyCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (konvoy *KonvoyCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return konvoy.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (konvoy *KonvoyCloudProvider) Refresh() error {
	if err := konvoy.konvoyManager.forceRefresh(); err != nil {
		return err
	}
	return nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (konvoy *KonvoyCloudProvider) Cleanup() error {
	return nil
}

// BuildKonvoy builds Konvoy cloud provider.
func BuildKonvoy(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	externalConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get kubeclient config for external cluster: %v", err)
	}

	scheme := runtime.NewScheme()
	if err := kommanderv1beta1.AddToScheme(scheme); err != nil {
		klog.Errorf("Unable to add konvoy management cluster to scheme: (%v)", err)
	}

	dynamicClient, err := client.New(externalConfig, client.Options{
		Scheme: scheme,
	})

	externalClient := kubeclient.NewForConfigOrDie(externalConfig)
	konvoyManager := &KonvoyManager{
		provisioner:   provisionerAWS,
		dynamicClient: dynamicClient,
		clusterName:   opts.ClusterName,
		kubeClient:    externalClient,
	}
	if err := konvoyManager.forceRefresh(); err != nil {
		klog.Fatalf("Failed to create Konovy Manager: %v", err)
	}

	provider, err := BuildKonvoyCloudProvider(konvoyManager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create Konvoy cloud provider: %v", err)
	}
	return provider
}
