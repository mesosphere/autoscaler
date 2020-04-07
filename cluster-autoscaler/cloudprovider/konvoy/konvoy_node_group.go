package konvoy

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// NodeGroup implements NodeGroup interface.
type NodeGroup struct {
	Name          string
	konvoyManager *KonvoyManager
	minSize       int
	maxSize       int
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
	nodeNames, err := nodeGroup.konvoyManager.GetNodeNamesForNodeGroup(nodeGroup.Name)
	if err != nil {
		return instances, err
	}
	for _, nodeName := range nodeNames {
		instances = append(instances, cloudprovider.Instance{Id: nodeName})
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
		nodeName := kubernetesNodeName(node, nodeGroup.konvoyManager.provisioner)
		if err := nodeGroup.konvoyManager.RemoveNodeFromNodeGroup(nodeGroup.Name, nodeName); err != nil {
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
	return nodeGroup.konvoyManager.setNodeGroupTargetSize(nodeGroup.Name, newSize)
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (nodeGroup *NodeGroup) TargetSize() (int, error) {
	size, err := nodeGroup.konvoyManager.GetNodeGroupTargetSize(nodeGroup.Name)
	klog.Infof("TargetSize() size: %v", size)
	return size, err
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
	return nodeGroup.konvoyManager.setNodeGroupTargetSize(nodeGroup.Name, newSize)
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

// Check that NodeGroup implements cloudprovider.NodeGroup interface
var _ cloudprovider.NodeGroup = (*NodeGroup)(nil)
