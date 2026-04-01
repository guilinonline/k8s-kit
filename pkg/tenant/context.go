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

import (
	"context"
)

// tenantKey is the unexported context key for tenant ID.
// Using an unexported type prevents collisions with context keys from other packages.
type tenantKey struct{}

// clusterKey is the unexported context key for cluster ID.
type clusterKey struct{}

// userKey is the unexported context key for user information.
type userKey struct{}

const (
	// DefaultTenant is the default tenant ID used when no tenant is specified.
	DefaultTenant = "default"
)

// WithTenant stores the tenant ID in the context.
// If tenantID is empty string, it defaults to DefaultTenant.
//
// Example:
//
//	ctx := tenant.WithTenant(context.Background(), "tenant-123")
func WithTenant(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		tenantID = DefaultTenant
	}
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// FromContext extracts the tenant ID from the context.
// If no tenant is set in the context, returns DefaultTenant.
//
// Example:
//
//	tenantID := tenant.FromContext(ctx)
func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tenantKey{}).(string); ok && v != "" {
		return v
	}
	return DefaultTenant
}

// WithCluster stores the cluster ID in the context.
// If clusterID is empty, returns the original context unchanged.
//
// Example:
//
//	ctx := tenant.WithCluster(context.Background(), "cluster-001")
func WithCluster(ctx context.Context, clusterID string) context.Context {
	if clusterID == "" {
		return ctx
	}
	return context.WithValue(ctx, clusterKey{}, clusterID)
}

// ClusterFromContext extracts the cluster ID from the context.
// If no cluster ID is set, returns empty string.
//
// Example:
//
//	clusterID := tenant.ClusterFromContext(ctx)
func ClusterFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(clusterKey{}).(string); ok {
		return v
	}
	return ""
}

// WithUser stores user information in the context.
// This is an optional convenience function for storing user identity.
//
// Example:
//
//	ctx := tenant.WithUser(context.Background(), "user-123", "John Doe")
func WithUser(ctx context.Context, userID, userName string) context.Context {
	if userID == "" {
		return ctx
	}
	return context.WithValue(ctx, userKey{}, map[string]string{
		"id":   userID,
		"name": userName,
	})
}

// UserFromContext extracts user information from the context.
// Returns userID, userName, and a boolean indicating if user info exists.
//
// Example:
//
//	userID, userName, ok := tenant.UserFromContext(ctx)
func UserFromContext(ctx context.Context) (userID, userName string, ok bool) {
	if v, ok := ctx.Value(userKey{}).(map[string]string); ok {
		return v["id"], v["name"], true
	}
	return "", "", false
}

// WithTenantAndCluster is a convenience function to set both tenant and cluster in one call.
//
// Example:
//
//	ctx := tenant.WithTenantAndCluster(context.Background(), "tenant-123", "cluster-001")
func WithTenantAndCluster(ctx context.Context, tenantID, clusterID string) context.Context {
	ctx = WithTenant(ctx, tenantID)
	ctx = WithCluster(ctx, clusterID)
	return ctx
}

// ExtractAll is a convenience function to extract both tenant and cluster at once.
//
// Example:
//
//	tenantID, clusterID := tenant.ExtractAll(ctx)
func ExtractAll(ctx context.Context) (tenantID, clusterID string) {
	return FromContext(ctx), ClusterFromContext(ctx)
}
