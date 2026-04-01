// Multi-Tenant Example
//
// This example demonstrates how to use the tenant package for multi-tenant scenarios.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/guilinonline/k8s-kit/pkg/client"
	"github.com/guilinonline/k8s-kit/pkg/cluster"
	"github.com/guilinonline/k8s-kit/pkg/tenant"
)

func main() {
	ctx := context.Background()

	// Create client factory
	clientFactory := client.NewFactory()

	// Create ClusterManager
	manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)

	// Register clusters with tenant info
	kubeconfig := []byte(`...`)
	manager.Register("cluster-001", kubeconfig, cluster.WithTenantID("tenant-001"))
	manager.Register("cluster-002", kubeconfig, cluster.WithTenantID("tenant-002"))

	// ============ Scenario 1: Without tenant context ============
	fmt.Println("=== Scenario 1: Get client without tenant context ===")
	cli, err := manager.GetClient("cluster-001")
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	_ = cli

	// ============ Scenario 2: With tenant context ============
	fmt.Println("=== Scenario 2: Get client with tenant context ===")

	// Create context with tenant and cluster
	ctx = tenant.WithTenant(ctx, "tenant-001")
	ctx = tenant.WithCluster(ctx, "cluster-001")

	// Use context to get client
	cli, err = manager.GetClientFromContext(ctx)
	if err != nil {
		log.Fatalf("Failed to get client from context: %v", err)
	}
	_ = cli

	// ============ Scenario 3: Using convenience function ============
	fmt.Println("=== Scenario 3: Using convenience function ===")
	ctx = tenant.WithTenantAndCluster(context.Background(), "tenant-002", "cluster-002")
	cli, err = manager.GetClientFromContext(ctx)
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	_ = cli

	// ============ Scenario 4: Extract all values ============
	fmt.Println("=== Scenario 4: Extract all values ===")
	tenantID, clusterID := tenant.ExtractAll(ctx)
	fmt.Printf("Tenant: %s, Cluster: %s\n", tenantID, clusterID)

	// ============ Scenario 5: HTTP middleware pattern ============
	fmt.Println("=== Scenario 5: Middleware pattern ===")
	// In a real HTTP service, you'd use middleware:
	// ctx = extractTenantFromRequest(r) // e.g., from JWT
	// ctx = extractClusterFromRequest(r) // e.g., from URL path
	// ctx = tenant.WithTenantAndCluster(ctx, tenantID, clusterID)
	// handler.ServeHTTP(w, r.WithContext(ctx))

	// Clean up
	manager.Stop()

	fmt.Println("Multi-tenant example completed!")
}
