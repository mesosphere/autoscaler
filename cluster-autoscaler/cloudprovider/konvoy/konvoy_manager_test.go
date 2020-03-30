package konvoy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	kubernetestesting "k8s.io/client-go/testing"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kclv1beta1 "github.com/mesosphere/kommander-cluster-lifecycle/clientapis/pkg/apis/kommander/v1beta1"
	konvoyclusterv1beta1 "github.com/mesosphere/kommander-cluster-lifecycle/clientapis/pkg/apis/kommander/v1beta1"
	konvoyv1beta1 "github.com/mesosphere/konvoy/clientapis/pkg/apis/konvoy/v1beta1"
)

func TestKonvoyManagerGetNodeGroups(t *testing.T) {

	var tests = []struct {
		description   string
		clusterName   string
		konvoyCluster *kclv1beta1.KonvoyCluster
		nodeGroups    []*NodeGroup
	}{
		{
			description:   "should return empty node groups",
			clusterName:   "test-cluster",
			nodeGroups:    nil,
			konvoyCluster: &kclv1beta1.KonvoyCluster{},
		},
		{
			description: "should return a node group",
			clusterName: "test-cluster",
			nodeGroups: []*NodeGroup{
				{
					Name:    "test-autoscaling-pool",
					minSize: 1,
					maxSize: 10,
				},
			},
			konvoyCluster: &kclv1beta1.KonvoyCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "kommander",
				},
				Spec: kclv1beta1.KonvoyClusterSpec{
					ProvisionerConfiguration: konvoyv1beta1.ClusterProvisionerSpec{
						NodePools: []konvoyv1beta1.MachinePool{
							{
								Name: "test-autoscaling-pool",
								AutoscalingOptions: &konvoyv1beta1.AutoscalingOptions{
									MinSize: utilpointer.Int32Ptr(1),
									MaxSize: utilpointer.Int32Ptr(10),
								},
							},
						},
					},
				},
			},
		},
		{
			description: "should skip node pool with autoscaling disabled",
			clusterName: "test-cluster",
			nodeGroups:  nil,
			konvoyCluster: &kclv1beta1.KonvoyCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "kommander",
				},
				Spec: kclv1beta1.KonvoyClusterSpec{
					ProvisionerConfiguration: konvoyv1beta1.ClusterProvisionerSpec{
						NodePools: []konvoyv1beta1.MachinePool{
							{
								Name: "test-pool",
							},
						},
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	err := konvoyclusterv1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	for _, test := range tests {
		dynamicClient := fake.NewFakeClientWithScheme(scheme, test.konvoyCluster)

		mgr := &KonvoyManager{
			provisioner:   "aws",
			dynamicClient: dynamicClient,
			clusterName:   test.clusterName,
		}
		mgr.forceRefresh()

		nodeGroups := mgr.GetNodeGroups()
		// set the konvoyManager to nil so we can easily compare the node groups
		for i := range nodeGroups {
			nodeGroups[i].konvoyManager = nil
		}
		assert.Equal(t, test.nodeGroups, nodeGroups, test.description)
	}
}

func TestKonvoyManagerGetNodeNamesForNodeGroup(t *testing.T) {

	var tests = []struct {
		description   string
		err           error
		expectedNodes []string
		node          *corev1.Node
		nodeGroup     string
		provisioner   string
	}{
		{
			description:   "should return empty node names list",
			nodeGroup:     "test-node-group",
			expectedNodes: []string{},
			node:          &corev1.Node{},
		},
		{
			description:   "should return list with a node name",
			nodeGroup:     "test-node-group",
			expectedNodes: []string{"test-node"},
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						nodeGroupLabel: "test-node-group",
					},
				},
			},
		},
		{
			description:   "should return list with node provider id",
			nodeGroup:     "test-node-group",
			expectedNodes: []string{"test-provider-id"},
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						nodeGroupLabel: "test-node-group",
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: "test-provider-id",
				},
			},
			provisioner: "aws",
		},
		{
			description:   "should return empty list because node does not have required label",
			nodeGroup:     "test-node-group",
			expectedNodes: []string{},
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo": "test-node-group",
					},
				},
			},
		},
		{
			description: "should return an error listing nodes",
			err:         fmt.Errorf("fake error"),
			node:        &corev1.Node{},
		},
	}

	scheme := runtime.NewScheme()
	err := konvoyclusterv1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	for _, test := range tests {

		kubeClient := kubernetesfake.NewSimpleClientset(test.node)

		mgr := &KonvoyManager{
			kubeClient:  kubeClient,
			provisioner: test.provisioner,
		}

		if test.err != nil {
			reactor := func(action kubernetestesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, test.err
			}
			kubeClient.PrependReactor("list", "nodes", reactor)
		}

		nodes, err := mgr.GetNodeNamesForNodeGroup(test.nodeGroup)

		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expectedNodes, nodes, test.description)
	}

}
