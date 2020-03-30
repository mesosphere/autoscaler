package konvoy

import (
	"context"
	"fmt"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kommanderv1beta1 "github.com/mesosphere/kommander-cluster-lifecycle/clientapis/pkg/apis/kommander/v1beta1"
)

const (
	nodeGroupLabel = "autoscaling.k8s.io/nodegroup"
	numRetries     = 3

	provisionerAWS   = "aws"
	provisionerAzure = "azure"

	unknownTargetSize = -1
)

type KonvoyManager struct {
	provisioner      string
	clusterName      string
	clusterNamespace string
	kubeClient       kubeclient.Interface
	dynamicClient    client.Client

	nodeGroupsMutex sync.RWMutex
	nodeGroups      []*NodeGroup
}

// GetNodeGroups returns all node groups configured for this cloud provider.
func (k *KonvoyManager) GetNodeGroups() []*NodeGroup {
	k.nodeGroupsMutex.RLock()
	defer k.nodeGroupsMutex.RUnlock()
	return k.nodeGroups
}

func (k *KonvoyManager) forceRefresh() error {
	k.nodeGroupsMutex.Lock()
	defer k.nodeGroupsMutex.Unlock()

	konvoyCluster := &kommanderv1beta1.KonvoyCluster{}
	clusterNamespacedName := types.NamespacedName{
		Namespace: k.clusterNamespace,
		Name:      k.clusterName,
	}
	err := k.dynamicClient.Get(context.Background(), clusterNamespacedName, konvoyCluster)
	if err != nil {
		klog.Errorf("Error retrieving the konvoy cluster: %v -- %v", konvoyCluster.Name, err)
		return err
	}

	var ngs []*NodeGroup
	for _, pool := range konvoyCluster.Spec.ProvisionerConfiguration.NodePools {
		// If autoscaling is enabled
		if pool.AutoscalingOptions != nil {
			ngs = append(ngs, &NodeGroup{
				minSize:       int(*pool.AutoscalingOptions.MinSize),
				maxSize:       int(*pool.AutoscalingOptions.MaxSize),
				Name:          pool.Name,
				konvoyManager: k,
			})
		}
	}

	k.nodeGroups = ngs
	return nil
}

// GetNodeNamesForNodeGroup returns a list of actual nodes in the cluster.
func (k *KonvoyManager) GetNodeNamesForNodeGroup(nodeGroup string) ([]string, error) {
	selector := labels.SelectorFromSet(labels.Set{nodeGroupLabel: nodeGroup})
	options := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	nodes, err := k.kubeClient.CoreV1().Nodes().List(options)
	if err != nil {
		klog.V(2).Infof("Error listing nodes")
		return nil, err
	}
	klog.V(2).Infof("List of nodes: %v", nodes)
	result := make([]string, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		result = append(result, kubernetesNodeName(&node, k.provisioner))
	}
	return result, nil
}

// kubernetesNodeName returns a node name that should be used based on the provider
func kubernetesNodeName(node *apiv1.Node, provisioner string) string {
	switch provisioner {
	case provisionerAWS, provisionerAzure:
		return node.Spec.ProviderID
	default:
		return node.ObjectMeta.Name
	}
}

// GetNodeGroupSize returns the current size for the node group as observed.
func (k *KonvoyManager) GetNodeGroupSize(nodeGroup string) (int, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{nodeGroupLabel: nodeGroup}))
	options := metav1.ListOptions{
		LabelSelector: selector.String(),
	}
	nodes, err := k.kubeClient.CoreV1().Nodes().List(options)
	if err != nil {
		klog.V(2).Infof("Error listing nodes")
		return 0, err
	}
	return len(nodes.Items), nil
}

// GetNodeGroupTargetSize returns the target size of the node group.
func (k *KonvoyManager) GetNodeGroupTargetSize(nodeGroupName string) (int, error) {
	konvoyCluster := &kommanderv1beta1.KonvoyCluster{}
	konvoyCluster.Name = k.clusterName
	clusterNamespacedName := types.NamespacedName{
		Namespace: k.clusterNamespace,
		Name:      konvoyCluster.Name,
	}
	if err := k.dynamicClient.Get(context.Background(), clusterNamespacedName, konvoyCluster); err != nil {
		return unknownTargetSize, err
	}

	for _, nodePool := range konvoyCluster.Spec.ProvisionerConfiguration.NodePools {
		if nodePool.Name == nodeGroupName {
			return int(nodePool.Count), nil
		}
	}

	return unknownTargetSize, fmt.Errorf("node group %s not found", nodeGroupName)
}

func (k *KonvoyManager) setNodeGroupTargetSize(nodeGroupName string, newSize int) error {
	var err error
	for i := 0; i < numRetries; i++ {
		konvoyCluster := &kommanderv1beta1.KonvoyCluster{}
		konvoyCluster.Name = k.clusterName
		clusterNamespacedName := types.NamespacedName{
			Namespace: k.clusterNamespace,
			Name:      konvoyCluster.Name,
		}
		err = k.dynamicClient.Get(context.Background(), clusterNamespacedName, konvoyCluster)
		if err != nil {
			klog.Warningf("Error retrieving the konvoy cluster: %v -- %v", konvoyCluster.Name, err)
		}

		targetPoolIndex := -1
		for i, pool := range konvoyCluster.Spec.ProvisionerConfiguration.NodePools {
			if pool.Name == nodeGroupName {
				targetPoolIndex = i
			}
		}

		if targetPoolIndex < 0 {
			return fmt.Errorf("node group %s does not exists", nodeGroupName)
		}

		konvoyCluster.Spec.ProvisionerConfiguration.NodePools[targetPoolIndex].Count = int32(newSize)

		if err = k.dynamicClient.Update(context.Background(), konvoyCluster); err != nil {
			klog.Errorf("Error updating the konvoy cluster %s: %v", konvoyCluster.Name, err)
			err = fmt.Errorf("failed to set target size %d for node group %s: %v", newSize, nodeGroupName, err)
		} else {
			klog.Infof("Konvoy %s cluster target size set to %d for node group %s", konvoyCluster.Name, newSize, nodeGroupName)
			return nil
		}
	}

	return err
}

// GetNodeGroupNameForNode returns the name of the node group to which the node
// belongs.
func (k *KonvoyManager) GetNodeGroupNameForNode(nodeName string) (string, error) {
	kubernetesNode, err := k.getNodeByName(nodeName)
	if kubernetesNode == nil || err != nil {
		return "", fmt.Errorf("node %s does not exist", nodeName)
	}
	nodeGroupName, ok := kubernetesNode.Labels[nodeGroupLabel]
	if ok {
		return nodeGroupName, nil
	}
	return "", fmt.Errorf("can't find nodegroup for node %s due to missing label %s", nodeName, nodeGroupLabel)
}

func (k *KonvoyManager) getNodeByName(name string) (*apiv1.Node, error) {
	result, err := k.kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Error listing nodes")
		return nil, err
	}
	//klog.V(2).Infof("List of nodes: %v", nodes)
	for _, node := range result.Items {
		nodeName := kubernetesNodeName(&node, k.provisioner)
		if nodeName == name {
			klog.V(2).Infof("Get node by name: %v", name)
			return &node, nil
		}
	}
	return nil, nil
}

// RemoveNodeFromNodeGroup marks given node for deletion and decreases target node group size
// by one.
func (k *KonvoyManager) RemoveNodeFromNodeGroup(nodeGroupName string, nodeName string) error {
	kubernetesNode, err := k.getNodeByName(nodeName)
	if kubernetesNode == nil || err != nil {
		klog.Warningf("Can't delete node %s from nodegroup %s. Node does not exist.", nodeName, nodeGroupName)
		return err
	}
	if kubernetesNode.ObjectMeta.Labels[nodeGroupLabel] != nodeGroupName {
		return fmt.Errorf("can't delete node %s from nodegroup %s. Node is not in nodegroup", nodeName, nodeGroupName)
	}

	currentTargetSize, err := k.GetNodeGroupTargetSize(nodeGroupName)
	if err != nil {
		return err
	}

	kubernetesNode.Annotations[KonvoyNodeAnnotationKey] = time.Now().String()
	if err := k.dynamicClient.Update(context.Background(), kubernetesNode); err != nil {
		return err
	}

	return k.setNodeGroupTargetSize(nodeGroupName, currentTargetSize-1)
}
