package konvoy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

		nodeGroups := mgr.GetNodeGroups()
		// set the konvoyManager to nil so we can easily compare the node groups
		for i := range nodeGroups {
			nodeGroups[i].konvoyManager = nil
		}
		assert.Equal(t, test.nodeGroups, nodeGroups, "%s: expected %v got %v", test.description, test.nodeGroups, nodeGroups)
	}
}
