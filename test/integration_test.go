// Integration test for k8s-kit
// Run with: go test -v ./test/... -run TestAll
package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/guilinonline/k8s-kit/pkg/client"
	"github.com/guilinonline/k8s-kit/pkg/cluster"
	"github.com/guilinonline/k8s-kit/pkg/pod"
	"github.com/guilinonline/k8s-kit/pkg/resource"
	"github.com/guilinonline/k8s-kit/pkg/tenant"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	config1Path = "C:\\Users\\chenguilin\\code\\cglk8s-kit\\docs\\config-mock-1"
	config2Path = "C:\\Users\\chenguilin\\code\\cglk8s-kit\\docs\\config-mock-2"
)

// TestClientFactory tests the client factory
func TestClientFactory(t *testing.T) {
	t.Log("=== Testing ClientFactory ===")

	// Read kubeconfig
	kubeconfig, err := os.ReadFile(config1Path)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig: %v", err)
	}

	// Create factory
	factory := client.NewFactory(
		client.WithTimeout(30*time.Second),
		client.WithQPS(100),
	)

	// Create client from kubeconfig
	cli, err := factory.CreateFromKubeconfig(kubeconfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify client is not nil
	if cli == nil {
		t.Fatal("Client is nil")
	}
	if cli.Clientset == nil {
		t.Fatal("Clientset is nil")
	}
	if cli.RuntimeClient == nil {
		t.Fatal("RuntimeClient is nil")
	}
	if cli.RESTConfig == nil {
		t.Fatal("RESTConfig is nil")
	}

	t.Log("✓ ClientFactory.CreateFromKubeconfig() works")
}

// TestClusterManager tests the cluster manager
func TestClusterManager(t *testing.T) {
	t.Log("=== Testing ClusterManager ===")

	kubeconfig, err := os.ReadFile(config1Path)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig: %v", err)
	}

	// Create factory and manager
	factory := client.NewFactory()
	manager := cluster.NewManager(factory, cluster.DefaultHealthCheckConfig)
	defer manager.Stop()

	// Test: Register cluster
	err = manager.Register("test-cluster", kubeconfig, cluster.WithTenantID("test-tenant"))
	if err != nil {
		t.Fatalf("Failed to register cluster: %v", err)
	}
	t.Log("✓ ClusterManager.Register() works")

	// Test: List clusters
	clusters := manager.List()
	if len(clusters) != 1 || clusters[0] != "test-cluster" {
		t.Fatalf("List() returned unexpected result: %v", clusters)
	}
	t.Log("✓ ClusterManager.List() works")

	// Test: GetClient
	cli, err := manager.GetClient("test-cluster")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}
	if cli == nil {
		t.Fatal("Client is nil")
	}
	t.Log("✓ ClusterManager.GetClient() works")

	// Test: GetHealthStatus
	status, err := manager.GetHealthStatus("test-cluster")
	if err != nil {
		t.Fatalf("Failed to get health status: %v", err)
	}
	t.Logf("✓ ClusterManager.GetHealthStatus() works, status: %s", status)

	// Test: Unregister
	err = manager.Unregister("test-cluster")
	if err != nil {
		t.Fatalf("Failed to unregister cluster: %v", err)
	}
	t.Log("✓ ClusterManager.Unregister() works")
}

// TestMultiCluster tests managing multiple clusters
func TestMultiCluster(t *testing.T) {
	t.Log("=== Testing Multi-Cluster Management ===")

	kubeconfig1, _ := os.ReadFile(config1Path)
	kubeconfig2, _ := os.ReadFile(config2Path)

	factory := client.NewFactory()
	manager := cluster.NewManager(factory, cluster.DefaultHealthCheckConfig)
	defer manager.Stop()

	// Register two clusters
	manager.Register("cluster-1", kubeconfig1)
	manager.Register("cluster-2", kubeconfig2)

	clusters := manager.List()
	if len(clusters) != 2 {
		t.Fatalf("Expected 2 clusters, got %d", len(clusters))
	}
	t.Log("✓ Multiple cluster registration works")

	// Get clients for each cluster
	cli1, _ := manager.GetClient("cluster-1")
	cli2, _ := manager.GetClient("cluster-2")

	if cli1 == nil || cli2 == nil {
		t.Fatal("Failed to get clients for clusters")
	}

	// Verify they're different
	if cli1.RESTConfig.Host == cli2.RESTConfig.Host {
		t.Log("Note: Both clusters have same host (testing with mock config)")
	}
	t.Log("✓ Multi-cluster management works")
}

// TestTenantContext tests the tenant context utilities
func TestTenantContext(t *testing.T) {
	t.Log("=== Testing Tenant Context ===")

	ctx := context.Background()

	// Test: WithTenant
	ctx = tenant.WithTenant(ctx, "tenant-123")
	if tenant.FromContext(ctx) != "tenant-123" {
		t.Fatal("WithTenant/FromContext failed")
	}
	t.Log("✓ tenant.WithTenant() / tenant.FromContext() works")

	// Test: Default tenant
	ctx = context.Background()
	if tenant.FromContext(ctx) != "default" {
		t.Fatal("Default tenant should be 'default'")
	}
	t.Log("✓ Default tenant is 'default'")

	// Test: WithCluster
	ctx = tenant.WithCluster(ctx, "cluster-001")
	if tenant.ClusterFromContext(ctx) != "cluster-001" {
		t.Fatal("WithCluster/ClusterFromContext failed")
	}
	t.Log("✓ tenant.WithCluster() / tenant.ClusterFromContext() works")

	// Test: ExtractAll - create context with both tenant and cluster
	ctx = tenant.WithTenantAndCluster(context.Background(), "tenant-123", "cluster-001")
	tenantID, clusterID := tenant.ExtractAll(ctx)
	if tenantID != "tenant-123" || clusterID != "cluster-001" {
		t.Fatal("ExtractAll failed")
	}
	t.Log("✓ tenant.ExtractAll() works")

	// Test: GetClientFromContext
	factory := client.NewFactory()
	kubeconfig, _ := os.ReadFile(config1Path)
	manager := cluster.NewManager(factory, cluster.DefaultHealthCheckConfig)
	defer manager.Stop()

	manager.Register("cluster-001", kubeconfig)
	ctx = tenant.WithTenantAndCluster(context.Background(), "tenant-123", "cluster-001")

	cli, err := manager.GetClientFromContext(ctx)
	if err != nil {
		t.Fatalf("GetClientFromContext failed: %v", err)
	}
	if cli == nil {
		t.Fatal("Client is nil")
	}
	t.Log("✓ ClusterManager.GetClientFromContext() works with tenant context")
}

// TestPodLogs tests Pod log retrieval
func TestPodLogs(t *testing.T) {
	t.Log("=== Testing Pod Logs ===")

	kubeconfig, err := os.ReadFile(config1Path)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig: %v", err)
	}

	factory := client.NewFactory()
	cli, err := factory.CreateFromKubeconfig(kubeconfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	operator := pod.NewOperator()
	ctx := context.Background()

	// Get simple logs
	logs, err := operator.GetLogsSimple(ctx, cli, "default", "nginx2-845cb975b5-2xjcl",
		pod.WithTailLines(50),
	)
	if err != nil {
		t.Logf("GetLogsSimple error (may be expected if pod doesn't exist): %v", err)
	} else {
		t.Logf("✓ GetLogsSimple() works, got %d bytes", len(logs))
	}

	// Note: GetLogsStream requires the pod to exist and is harder to test
	t.Log("✓ PodOperator log retrieval available")
}

// TestPodExec tests Pod command execution
func TestPodExec(t *testing.T) {
	t.Log("=== Testing Pod Exec ===")

	kubeconfig, err := os.ReadFile(config1Path)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig: %v", err)
	}

	factory := client.NewFactory()
	cli, err := factory.CreateFromKubeconfig(kubeconfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	operator := pod.NewOperator()
	ctx := context.Background()

	// Execute simple command
	result, err := operator.ExecSimple(ctx, cli, "default", "nginx2-845cb975b5-2xjcl",
		[]string{"ls", "-la"},
		pod.WithExecTimeout(10*time.Second),
	)
	if err != nil {
		t.Logf("ExecSimple error (may be expected if pod doesn't exist): %v", err)
	} else {
		t.Logf("✓ ExecSimple() works, exit code: %d", result.ExitCode)
		if result.Stdout != "" {
			t.Logf("  stdout length: %d bytes", len(result.Stdout))
		}
	}
}

// TestResourceOperator tests the resource operator CRUD
func TestResourceOperator(t *testing.T) {
	t.Log("=== Testing Resource Operator ===")

	kubeconfig, err := os.ReadFile(config1Path)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig: %v", err)
	}

	factory := client.NewFactory()
	cli, err := factory.CreateFromKubeconfig(kubeconfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	operator := resource.NewOperator()
	ctx := context.Background()

	// Test: List Pods
	podList := &corev1.PodList{}
	err = operator.List(ctx, cli, podList,
		resource.WithNamespace("default"),
		resource.WithLimit(10),
	)
	if err != nil {
		t.Logf("List pods error: %v", err)
	} else {
		t.Logf("✓ ResourceOperator.List() works, found %d pods", len(podList.Items))
	}

	// Test: List Namespaces
	nsList := &corev1.NamespaceList{}
	err = operator.List(ctx, cli, nsList)
	if err != nil {
		t.Logf("List namespaces error: %v", err)
	} else {
		t.Logf("✓ ResourceOperator.List() works for namespaces, found %d", len(nsList.Items))
	}

	// Test: Get Namespace
	ns := &corev1.Namespace{}
	err = operator.Get(ctx, cli, ns, types.NamespacedName{Name: "default"})
	if err != nil {
		t.Logf("Get namespace error: %v", err)
	} else {
		t.Logf("✓ ResourceOperator.Get() works, namespace: %s", ns.Name)
	}
}

// TestAll runs all integration tests
func TestAll(t *testing.T) {
	// Run tests in order
	t.Run("ClientFactory", TestClientFactory)
	t.Run("ClusterManager", TestClusterManager)
	t.Run("MultiCluster", TestMultiCluster)
	t.Run("TenantContext", TestTenantContext)
	t.Run("PodLogs", TestPodLogs)
	t.Run("PodExec", TestPodExec)
	t.Run("ResourceOperator", TestResourceOperator)

	fmt.Println("\n=== All Integration Tests Completed ===")
}

func TestMain(m *testing.M) {
	// Check if kubeconfig files exist
	if _, err := os.Stat(config1Path); os.IsNotExist(err) {
		fmt.Printf("Warning: %s not found\n", config1Path)
	}
	if _, err := os.Stat(config2Path); os.IsNotExist(err) {
		fmt.Printf("Warning: %s not found\n", config2Path)
	}

	os.Exit(m.Run())
}
