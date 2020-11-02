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

	konvoyautoprovv1beta1 "github.com/mesosphere/konvoy/auto-provisioning/apis/pkg/apis/kommander/v1beta1"
	konvoyv1beta1 "github.com/mesosphere/konvoy/clientapis/pkg/apis/konvoy/v1beta1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

func TestKonvoyManagerGetNodeGroups(t *testing.T) {

	var tests = []struct {
		description   string
		clusterName   string
		konvoyCluster *konvoyautoprovv1beta1.KonvoyCluster
		nodeGroups    []*NodeGroup
	}{
		{
			description:   "should return empty node groups",
			clusterName:   "test-cluster",
			nodeGroups:    nil,
			konvoyCluster: &konvoyautoprovv1beta1.KonvoyCluster{},
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
			konvoyCluster: &konvoyautoprovv1beta1.KonvoyCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "kommander",
				},
				Spec: konvoyautoprovv1beta1.KonvoyClusterSpec{
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
			konvoyCluster: &konvoyautoprovv1beta1.KonvoyCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "kommander",
				},
				Spec: konvoyautoprovv1beta1.KonvoyClusterSpec{
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
	err := konvoyautoprovv1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	for _, test := range tests {
		dynamicClient := fake.NewFakeClientWithScheme(scheme, test.konvoyCluster)

		kubeClient := kubernetesfake.NewSimpleClientset()

		kubeEventRecorder := kube_util.CreateEventRecorderWithScheme(kubeClient, scheme)

		mgr := &KonvoyManager{
			kubeClient:    kubeClient,
			dynamicClient: dynamicClient,
			clusterName:   test.clusterName,
			eventRecorder: kubeEventRecorder,
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

func TestKonvoyManagerSetTargetSizeIgnored(t *testing.T) {

	var tests = []struct {
		description   string
		clusterName   string
		konvoyCluster *konvoyautoprovv1beta1.KonvoyCluster
		nodeGroups    []*NodeGroup
	}{
		{
			description: "should return nil because cluster is paused",
			clusterName: "test-cluster",
			nodeGroups: []*NodeGroup{
				{
					Name:    "test-autoscaling-pool",
					minSize: 1,
					maxSize: 10,
				},
			},
			konvoyCluster: &konvoyautoprovv1beta1.KonvoyCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "kommander",
				},
				Spec: konvoyautoprovv1beta1.KonvoyClusterSpec{
					ProvisioningPaused: true,
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
	}

	scheme := runtime.NewScheme()
	err := konvoyautoprovv1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	for _, test := range tests {
		dynamicClient := fake.NewFakeClientWithScheme(scheme, test.konvoyCluster)

		kubeClient := kubernetesfake.NewSimpleClientset()

		kubeEventRecorder := kube_util.CreateEventRecorderWithScheme(kubeClient, scheme)

		mgr := &KonvoyManager{
			dynamicClient: dynamicClient,
			clusterName:   test.clusterName,
			eventRecorder: kubeEventRecorder,
		}
		mgr.forceRefresh()

		nodeGroups := mgr.GetNodeGroups()
		// set the konvoyManager to nil so we can easily compare the node groups
		for i := range nodeGroups {
			nodeGroups[i].konvoyManager = nil
		}
		assert.Equal(t, test.nodeGroups, nodeGroups, test.description)

		err = mgr.setNodeGroupTargetSize("test-autoscaling-pool", 5)
		assert.Error(t, err)
	}
}

func TestKonvoyManagerGetNodeNamesForNodeGroup(t *testing.T) {

	var tests = []struct {
		description   string
		err           error
		expectedNodes []string
		node          *corev1.Node
		nodeGroup     string
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
	err := konvoyautoprovv1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	for _, test := range tests {

		kubeClient := kubernetesfake.NewSimpleClientset(test.node)

		kubeEventRecorder := kube_util.CreateEventRecorderWithScheme(kubeClient, scheme)

		mgr := &KonvoyManager{
			kubeClient:  kubeClient,
			eventRecorder: kubeEventRecorder,
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
