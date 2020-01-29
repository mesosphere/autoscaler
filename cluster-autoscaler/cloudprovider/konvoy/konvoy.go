package konvoy

import (
	"fmt"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	konvoyclusterv1beta1 "github.com/mesosphere/kommander-cluster-lifecycle/pkg/apis/kommander/v1beta1"

	"k8s.io/klog"
)

const (
	// ProviderName is the cloud provider name for konvoy
	ProviderName = "konvoy"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "konvoy.d2iq.com/gpu"
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
	konvoyManager      *KonvoyManager
	nodeGroups         []*NodeGroup
	resourceLimiter    *cloudprovider.ResourceLimiter
}

// BuildKonvoyCloudProvider builds a CloudProvider for konvoy. Builds
// node groups from passed in specs.
func BuildKonvoyCloudProvider(konvoyManager *KonvoyManager, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*KonvoyCloudProvider, error) {
	konvoy := &KonvoyCloudProvider{
		konvoyManager:      konvoyManager,
		nodeGroups:         make([]*NodeGroup, 0),
		resourceLimiter:    resourceLimiter,
	}
	for _, spec := range specs {
		if err := konvoy.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}
	return konvoy, nil
}

func (konvoy *KonvoyCloudProvider) addNodeGroup(spec string) error {
	nodeGroup, err := buildNodeGroup(spec, konvoy.konvoyManager)
	if err != nil {
		return err
	}
	klog.V(2).Infof("adding node group: %s", nodeGroup.Name)
	konvoy.nodeGroups = append(konvoy.nodeGroups, nodeGroup)
	return nil
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
	result := make([]cloudprovider.NodeGroup, 0, len(konvoy.nodeGroups))
	for _, nodegroup := range konvoy.nodeGroups {
		result = append(result, nodegroup)
	}
	return result
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (konvoy *KonvoyCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}




// NodeGroupForNode returns the node group for the given node.
func (konvoy *KonvoyCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	nodeGroupName, err := konvoy.konvoyManager.GetNodeGroupForNode(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	for _, nodeGroup := range konvoy.nodeGroups {
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
	return nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (konvoy *KonvoyCloudProvider) Cleanup() error {
	return nil
}

// NodeGroup implements NodeGroup interface.
type NodeGroup struct {
	Name               string
	konvoyManager *KonvoyManager
	minSize            int
	maxSize            int
}

// Id returns nodegroup name.
func (nodeGroup *NodeGroup) Id() string {
	return nodeGroup.Name
}

// MinSize returns minimum size of the node group.
func (nodeGroup *NodeGroup) MinSize() int {
	return nodeGroup.minSize
}

// MaxSize returns maximum size of the node group.
func (nodeGroup *NodeGroup) MaxSize() int {
	return nodeGroup.maxSize
}

// Debug returns a debug string for the nodegroup.
func (nodeGroup *NodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", nodeGroup.Id(), nodeGroup.MinSize(), nodeGroup.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (nodeGroup *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0)
	nodes, err := nodeGroup.konvoyManager.GetNodeNamesForNodeGroup(nodeGroup.Name)
	if err != nil {
		return instances, err
	}
	for _, node := range nodes {
		instances = append(instances, cloudprovider.Instance{Id: "" + node})
	}
	return instances, nil
}

// DeleteNodes deletes the specified nodes from the node group.
func (nodeGroup *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.Infof("DeleteNodes %v", nodes)
	size, err := nodeGroup.konvoyManager.GetNodeGroupTargetSize(nodeGroup.Name)
	if err != nil {
		return err
	}
	if size <= nodeGroup.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	for _, node := range nodes {
		// FIXME: aws based
		if err := nodeGroup.konvoyManager.RemoveNodeFromNodeGroup(nodeGroup.Name, node.Spec.ProviderID); err != nil {
			return err
		}
	}
	return nil
}

// IncreaseSize increases NodeGroup size.
func (nodeGroup *NodeGroup) IncreaseSize(delta int) error {
	klog.Infof("IncreaseSize %v", delta)
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := nodeGroup.konvoyManager.GetNodeGroupTargetSize(nodeGroup.Name)
	if err != nil {
		return err
	}
	newSize := int(size) + delta
	if newSize > nodeGroup.MaxSize() {
		return fmt.Errorf("size increase too large, desired: %d max: %d", newSize, nodeGroup.MaxSize())
	}
	return nodeGroup.konvoyManager.SetNodeGroupSize(nodeGroup.Name, newSize)
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (nodeGroup *NodeGroup) TargetSize() (int, error) {
	size, err := nodeGroup.konvoyManager.GetNodeGroupTargetSize(nodeGroup.Name)
	klog.Infof("TargetSize() size: %v", size)
	return int(size), err
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (nodeGroup *NodeGroup) DecreaseTargetSize(delta int) error {
	klog.Infof("DecreaseTargetSize %v", delta)
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size, err := nodeGroup.konvoyManager.GetNodeGroupTargetSize(nodeGroup.Name)
	if err != nil {
		return err
	}
	nodes, err := nodeGroup.konvoyManager.GetNodeNamesForNodeGroup(nodeGroup.Name)
	if err != nil {
		return err
	}
	newSize := int(size) + delta
	if newSize < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes, targetSize: %d delta: %d existingNodes: %d",
			size, delta, len(nodes))
	}
	return nodeGroup.konvoyManager.SetNodeGroupSize(nodeGroup.Name, newSize)
}

// TemplateNodeInfo returns a node template for this node group.
func (nodeGroup *NodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side.
func (nodeGroup *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (nodeGroup *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (nodeGroup *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (nodeGroup *NodeGroup) Autoprovisioned() bool {
	return false
}

func buildNodeGroup(value string, konvoyManager *KonvoyManager) (*NodeGroup, error) {
	spec, err := dynamic.SpecFromString(value, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	klog.Infof("buildNodeGroup: value %v; %v; %v; %v", value, spec.MinSize, spec.MaxSize, spec.Name)

	nodeGroup := &NodeGroup{
		Name:               spec.Name,
		konvoyManager: konvoyManager,
		minSize:            spec.MinSize,
		maxSize:            spec.MaxSize,
	}

	return nodeGroup, nil
}

// BuildKonvoy builds Konvoy cloud provider.
func BuildKonvoy(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	externalConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get kubeclient config for external cluster: %v", err)
	}

	/*konvoyConfig, err := clientcmd.BuildConfigFromFlags("", "/kubeconfig/cluster_autoscaler.kubeconfig")
	if err != nil {
		klog.Fatalf("Failed to get kubeclient config for konvoy cluster: %v", err)
	}*/

	//stop := make(chan struct{})

	//Add route Openshift scheme
	scheme := runtime.NewScheme()
	if err := konvoyclusterv1beta1.AddToScheme(scheme); err != nil {
		klog.Errorf("Unable to add konvoy management cluster to scheme: (%v)", err)
	}

	dynamicClient, err := client.New(externalConfig, client.Options{
		Scheme: scheme,
	})

	externalClient := kubeclient.NewForConfigOrDie(externalConfig)
  konvoyManager := &KonvoyManager{
		provisioner: "aws",
		dynamicClient: dynamicClient,
    clusterName: opts.ClusterName,
    kubeClient: externalClient,
    createNodeQueue:        make(chan string, 1000),
    nodeGroupQueueSize:     make(map[string]int),
    nodeGroupQueueSizeLock: sync.Mutex{},
  }
	//konvoyClient := kubeclient.NewForConfigOrDie(konvoyConfig)

	/*externalInformerFactory := informers.NewSharedInformerFactory(externalClient, 0)
	konvoyInformerFactory := informers.NewSharedInformerFactory(konvoyClient, 0)
	konvoyNodeInformer := konvoyInformerFactory.Core().V1().Nodes()
	go konvoyNodeInformer.Informer().Run(stop)

	kubemarkController, err := kubemark.NewKubemarkController(externalClient, externalInformerFactory,
		konvoyClient, konvoyNodeInformer)
	if err != nil {
		klog.Fatalf("Failed to create Konvoy cloud provider: %v", err)
	}

	externalInformerFactory.Start(stop)
	if !kubemarkController.WaitForCacheSync(stop) {
		klog.Fatalf("Failed to sync caches for konvoy controller")
	}
	go kubemarkController.Run(stop)*/

	provider, err := BuildKonvoyCloudProvider(konvoyManager, do.NodeGroupSpecs, rl)
	if err != nil {
		klog.Fatalf("Failed to create Konvoy cloud provider: %v", err)
	}
	return provider
}
