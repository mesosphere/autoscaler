package konvoy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	konvoyv1beta1 "github.com/mesosphere/konvoy/clientapis/pkg/apis/konvoy/v1beta1"
	konvoyclusterv1beta1 "github.com/mesosphere/yakcl/clientapis/pkg/apis/kommander/v1beta1"
	yakclv1beta1 "github.com/mesosphere/yakcl/clientapis/pkg/apis/kommander/v1beta1"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
)

var testKonvoyCluster = &yakclv1beta1.KonvoyCluster{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-cluster",
		Namespace: "kommander",
	},
	Spec: yakclv1beta1.KonvoyClusterSpec{
		ProvisioningPaused: false,
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
}

func TestKonvoyNodeGroupForNode(t *testing.T) {

	var tests = []struct {
		description   string
		expectedError bool
		node          *corev1.Node
		nodeGroup     NodeGroup
	}{
		{
			description:   "should return nil node group due to a master node",
			expectedError: false,
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						KubernetesMasterNodeLabel: "master",
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: "test-provider-id",
				},
			},
		},
		{
			description: "should return the group name",
			nodeGroup: NodeGroup{
				Name: "test-node-group",
			},
			expectedError: false,
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
			description:   "should return a nil nodeGroup due to a missing nodegroup label",
			expectedError: false,
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo": "test-node-group",
					},
				},
				Spec: corev1.NodeSpec{
					ProviderID: "test-provider-id",
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	err := konvoyclusterv1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	for _, test := range tests {
		dynamicClient := fake.NewFakeClientWithScheme(scheme, testKonvoyCluster)

		kubeClient := kubernetesfake.NewSimpleClientset(test.node)
		kubeEventRecorder := kube_util.CreateEventRecorderWithScheme(kubeClient, scheme)

		mgr := &KonvoyManager{
			dynamicClient: dynamicClient,
			clusterName:   "test-cluster",
			kubeClient:    kubeClient,
			eventRecorder: kubeEventRecorder,
		}

		mgr.forceRefresh()

		konvoyCloudProvider := &KonvoyCloudProvider{
			konvoyManager: mgr,
		}

		nodeGroupName, err := konvoyCloudProvider.NodeGroupForNode(test.node)
		if nodeGroupName != nil {
			assert.Equal(t, test.nodeGroup.Name, nodeGroupName)
		} else {
			// check that it was expected to be nil
			assert.Equal(t, nil, nodeGroupName)
		}
		assert.Equal(t, test.expectedError, (err != nil))
	}
}
