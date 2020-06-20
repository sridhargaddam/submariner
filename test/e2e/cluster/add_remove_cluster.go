package cluster

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sridhargaddam/shipyard/test/e2e/framework"
	"github.com/sridhargaddam/shipyard/test/e2e/tcp"
)

var _ = PDescribe("[expansion] Test expanding/shrinking an existing cluster fleet", func() {
	f := framework.NewFramework("add-remove-cluster")

	It("Should be able to add and remove third cluster", func() {
		clusterAName := framework.TestContext.ClusterIDs[framework.ClusterA]
		clusterBName := framework.TestContext.ClusterIDs[framework.ClusterB]
		clusterCName := framework.TestContext.ClusterIDs[framework.ClusterC]

		By(fmt.Sprintf("Verifying no GW nodes are present on cluster %q", clusterCName))
		gatewayNode := f.FindNodesByGatewayLabel(framework.ClusterC, true)
		Expect(gatewayNode).To(HaveLen(0), fmt.Sprintf("Expected no gateway node on %q", framework.ClusterC))

		By(fmt.Sprintf("Verifying that a pod in cluster %q cannot connect to a pod in cluster %q", clusterAName, clusterCName))
		tcp.RunNoConnectivityTest(tcp.ConnectivityTestParams{
			Framework:             f,
			FromCluster:           framework.ClusterA,
			FromClusterScheduling: framework.GatewayNode,
			ToCluster:             framework.ClusterC,
			ToClusterScheduling:   framework.NonGatewayNode,
		})

		By(fmt.Sprintf("Verifying that a pod in cluster %q cannot connect to a service in cluster %q", clusterBName, clusterCName))
		tcp.RunNoConnectivityTest(tcp.ConnectivityTestParams{
			Framework:             f,
			ToEndpointType:        tcp.ServiceIP,
			FromCluster:           framework.ClusterB,
			FromClusterScheduling: framework.NonGatewayNode,
			ToCluster:             framework.ClusterC,
			ToClusterScheduling:   framework.NonGatewayNode,
		})

		nonGatewayNodes := f.FindNodesByGatewayLabel(framework.ClusterC, false)
		Expect(nonGatewayNodes).ToNot(HaveLen(0), fmt.Sprintf("No non-gateway nodes found on %q", clusterCName))
		nonGatewayNode := nonGatewayNodes[0].Name
		By(fmt.Sprintf("Adding cluster %q by setting the gateway label on node %q", clusterCName, nonGatewayNode))
		f.SetGatewayLabelOnNode(framework.ClusterC, nonGatewayNode, true)

		enginePod := f.AwaitSubmarinerEnginePod(framework.ClusterC)
		By(fmt.Sprintf("Found submariner engine pod %q on %q", enginePod.Name, clusterCName))

		By(fmt.Sprintf("Checking connectivity between clusters"))
		tcp.RunConnectivityTest(tcp.ConnectivityTestParams{
			Framework:             f,
			FromCluster:           framework.ClusterB,
			FromClusterScheduling: framework.GatewayNode,
			ToCluster:             framework.ClusterC,
			ToClusterScheduling:   framework.GatewayNode,
		})

		tcp.RunConnectivityTest(tcp.ConnectivityTestParams{
			Framework:             f,
			ToEndpointType:        tcp.ServiceIP,
			FromCluster:           framework.ClusterA,
			FromClusterScheduling: framework.NonGatewayNode,
			ToCluster:             framework.ClusterC,
			ToClusterScheduling:   framework.NonGatewayNode,
		})

		By(fmt.Sprintf("Removing cluster %q by unsetting the gateway label and deleting submariner engine pod %q", clusterCName, enginePod.Name))
		f.SetGatewayLabelOnNode(framework.ClusterC, nonGatewayNode, false)
		f.DeletePod(framework.ClusterC, enginePod.Name, framework.TestContext.SubmarinerNamespace)

		By(fmt.Sprintf("Verifying that a pod in cluster %q cannot connect to a service in cluster %q", clusterAName, clusterCName))
		tcp.RunNoConnectivityTest(tcp.ConnectivityTestParams{
			Framework:             f,
			FromCluster:           framework.ClusterA,
			FromClusterScheduling: framework.GatewayNode,
			ToCluster:             framework.ClusterC,
			ToClusterScheduling:   framework.NonGatewayNode,
		})

		By(fmt.Sprintf("Verifying that a pod in cluster %q cannot connect to a pod in cluster %q", clusterBName, clusterCName))
		tcp.RunNoConnectivityTest(tcp.ConnectivityTestParams{
			Framework:             f,
			ToEndpointType:        tcp.ServiceIP,
			FromCluster:           framework.ClusterB,
			FromClusterScheduling: framework.NonGatewayNode,
			ToCluster:             framework.ClusterC,
			ToClusterScheduling:   framework.NonGatewayNode,
		})
	})
})
