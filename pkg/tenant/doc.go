// Package tenant provides multi-tenancy support through Context-based tenant information passing.
//
// This package implements non-intrusive tenant identification and context propagation.
// It does not implement business-level tenant isolation (like network isolation, resource quotas),
// but provides the infrastructure for tenant identification and context passing.
//
// Example usage:
//
//	// Inject tenant into context
//	ctx := tenant.WithTenant(context.Background(), "tenant-123")
//	ctx = tenant.WithCluster(ctx, "cluster-001")
//
//	// Extract tenant from context
//	tenantID := tenant.FromContext(ctx)  // returns "tenant-123"
//	clusterID := tenant.ClusterFromContext(ctx)  // returns "cluster-001"
//
//	// Default tenant when not set
//	ctx := context.Background()
//	tenantID := tenant.FromContext(ctx)  // returns "default"
package tenant
