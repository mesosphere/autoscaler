package konvoy

import (
  "context"
	"fmt"
  "sync"

	apiv1 "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/types"
  "k8s.io/apimachinery/pkg/labels"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
  "sigs.k8s.io/controller-runtime/pkg/client"

  konvoyclusterv1beta1 "github.com/mesosphere/kommander-cluster-lifecycle/pkg/apis/kommander/v1beta1"
  konvoyv1beta1  "github.com/mesosphere/konvoy/pkg/apis/konvoy/v1beta1"
)

const (
  nodeGroupLabel    = "autoscaling.k8s.io/nodegroup"
  numRetries        = 3
)

type KonvoyManager struct {
  provisioner string
  clusterName            string
  kubeClient kubeclient.Interface
  createNodeQueue        chan string
  nodeGroupQueueSize     map[string]int
  nodeGroupQueueSizeLock sync.Mutex
  dynamicClient client.Client
}

// GetNodeGroupSize returns the current size for the node group as observed.
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
    if k.provisioner == "aws" {
		  result = append(result, node.Spec.ProviderID)
    } else {
      result = append(result, node.ObjectMeta.Name)
    }
	}
	return result, nil
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

// GetNodeGroupTargetSize returns the size of the node group as a sum of current
// observed size and number of upcoming nodes.
func (k *KonvoyManager) GetNodeGroupTargetSize(nodeGroup string) (int, error) {
	k.nodeGroupQueueSizeLock.Lock()
	defer k.nodeGroupQueueSizeLock.Unlock()
	realSize, err := k.GetNodeGroupSize(nodeGroup)
	if err != nil {
		return realSize, err
	}
	return realSize + k.nodeGroupQueueSize[nodeGroup], nil
}

// SetNodeGroupSize changes the size of node group by adding or removing nodes.
func (k *KonvoyManager) SetNodeGroupSize(nodeGroup string, size int) error {
	currSize, err := k.GetNodeGroupTargetSize(nodeGroup)
	if err != nil {
		return err
	}
	switch delta := size - currSize; {
	case delta < 0:
		absDelta := -delta
		nodes, err := k.GetNodeNamesForNodeGroup(nodeGroup)
		if err != nil {
			return err
		}
		if len(nodes) < absDelta {
			return fmt.Errorf("can't remove %d nodes from %s nodegroup, not enough nodes: %d", absDelta, nodeGroup, len(nodes))
		}
		for i, node := range nodes {
			if i == absDelta {
				return nil
			}
			if err := k.RemoveNodeFromNodeGroup(nodeGroup, node); err != nil {
				return err
			}
		}
	case delta > 0:
		k.nodeGroupQueueSizeLock.Lock()
    defer k.nodeGroupQueueSizeLock.Unlock()
		for i := 0; i < delta; i++ {
			k.nodeGroupQueueSize[nodeGroup]++
			if err := k.addNodeToNodeGroup(nodeGroup); err != nil {
        k.nodeGroupQueueSize[nodeGroup]--
        return err
      }
		}
	}

	return nil
}

func (k *KonvoyManager) addNodeToNodeGroup(nodeGroup string) error {
	var err error
	for i := 0; i < numRetries; i++ {
    konvoyCluster := &konvoyclusterv1beta1.KonvoyManagementCluster{}
    konvoyCluster.Name = k.clusterName
    clusterNamespacedName := types.NamespacedName{
      Namespace: "kommander",
      Name:      konvoyCluster.Name,
    }
  	err = k.dynamicClient.Get(context.Background(), clusterNamespacedName, konvoyCluster)
  	if err != nil {
      klog.Warningf("Error retrieving the konvoy cluster")
  	}
    newPool := make([]konvoyv1beta1.MachinePool, len(konvoyCluster.Spec.ProvisionerConfiguration.NodePools))
    for _, pool := range konvoyCluster.Spec.ProvisionerConfiguration.NodePools {
      if pool.Name == nodeGroup {
        pool.Count++
      }
      newPool = append(newPool, pool)
    }
    konvoyCluster.Spec.ProvisionerConfiguration.NodePools = newPool

    if err = k.dynamicClient.Update(context.Background(), konvoyCluster); err != nil {
      klog.Warningf("Error updating the konvoy cluster")
      err = fmt.Errorf("Failed to add node to group %s: %v", nodeGroup, err)
    } else {
      return nil
    }
	}

	return err
}

// GetNodeGroupForNode returns the name of the node group to which the node
// belongs.
func (k *KonvoyManager) GetNodeGroupForNode(node string) (string, error) {
	konvoyNode, err := k.getNodeByName(node)
	if konvoyNode == nil || err != nil {
		return "", fmt.Errorf("node %s does not exist", node)
	}
	nodeGroup, ok := konvoyNode.Labels[nodeGroupLabel]
	if ok {
		return nodeGroup, nil
	}
	return "", fmt.Errorf("can't find nodegroup for node %s due to missing label %s", node, nodeGroupLabel)
}

func (k *KonvoyManager) getNodeByName(name string) (*apiv1.Node, error) {
  nodes, err := k.kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
  if err != nil {
    klog.Warningf("Error listing nodes")
		return nil, err
	}
  //klog.V(2).Infof("List of nodes: %v", nodes)
	for _, node := range nodes.Items {
    if k.provisioner == "aws" {
      if node.Spec.ProviderID == name {
        klog.V(2).Infof("Get node by name: %v", name)
        return &node, nil
      }
    } else {
      if node.Name == name {
        klog.V(2).Infof("Get node by name: %v", name)
        return &node, nil
      }
    }
	}
	return nil, nil
}

func (k *KonvoyManager) RemoveNodeFromNodeGroup(nodeGroup string, node string) error {
	konvoyNode, err := k.getNodeByName(node)
	if konvoyNode == nil || err != nil {
		klog.Warningf("Can't delete node %s from nodegroup %s. Node does not exist.", node, nodeGroup)
		return nil
	}
	if konvoyNode.ObjectMeta.Labels[nodeGroupLabel] != nodeGroup {
		return fmt.Errorf("can't delete node %s from nodegroup %s. Node is not in nodegroup", node, nodeGroup)
	}

  /*
  TODO: we could mark a node for deletion or delete it using Konvoy library
  if err := k.kubeClient.CoreV1().Nodes().Delete(node.Name, &metav1.DeleteOptions{}); err != nil {
    klog.Errorf("failed to delete node %s from Konvoy cluster, err: %v", node.Name, err)
  }
  */
  for i := 0; i < numRetries; i++ {
    // Decrease the size of the pool in Konvoy
    konvoyCluster := &konvoyclusterv1beta1.KonvoyManagementCluster{}
    konvoyCluster.Name = k.clusterName
    clusterNamespacedName := types.NamespacedName{
			Namespace: "kommander",
			Name:      konvoyCluster.Name,
		}
  	err = k.dynamicClient.Get(context.Background(), clusterNamespacedName, konvoyCluster)
  	if err != nil {
      klog.Warningf("Error retrieving the konvoy cluster")
  	}
    newPool := make([]konvoyv1beta1.MachinePool, len(konvoyCluster.Spec.ProvisionerConfiguration.NodePools))
    for _, pool := range konvoyCluster.Spec.ProvisionerConfiguration.NodePools {
      if pool.Name == nodeGroup {
        pool.Count--
      }
      newPool = append(newPool, pool)
    }
    konvoyCluster.Spec.ProvisionerConfiguration.NodePools = newPool

    if err = k.dynamicClient.Update(context.Background(), konvoyCluster); err != nil {
      klog.Warningf("Error updating the konvoy cluster")
      err = fmt.Errorf("Failed to delete node %s: %v", node, err)
    } else {
      return nil
    }
  }
	return err
}
