package tenant

import (
	"context"
	"testing"
)

func TestWithTenant(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenant(ctx, "tenant-123")

	if got := FromContext(ctx); got != "tenant-123" {
		t.Errorf("Expected tenant-123, got %s", got)
	}
}

func TestDefaultTenant(t *testing.T) {
	ctx := context.Background()

	if got := FromContext(ctx); got != DefaultTenant {
		t.Errorf("Expected default, got %s", got)
	}
}

func TestEmptyTenantDefaultsToDefault(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenant(ctx, "")

	if got := FromContext(ctx); got != DefaultTenant {
		t.Errorf("Expected default, got %s", got)
	}
}

func TestWithCluster(t *testing.T) {
	ctx := context.Background()
	ctx = WithCluster(ctx, "cluster-001")

	if got := ClusterFromContext(ctx); got != "cluster-001" {
		t.Errorf("Expected cluster-001, got %s", got)
	}
}

func TestEmptyClusterReturnsEmpty(t *testing.T) {
	ctx := context.Background()

	if got := ClusterFromContext(ctx); got != "" {
		t.Errorf("Expected empty string, got %s", got)
	}
}

func TestWithTenantAndCluster(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenantAndCluster(ctx, "tenant-123", "cluster-001")

	tenantID, clusterID := ExtractAll(ctx)

	if tenantID != "tenant-123" {
		t.Errorf("Expected tenant-123, got %s", tenantID)
	}
	if clusterID != "cluster-001" {
		t.Errorf("Expected cluster-001, got %s", clusterID)
	}
}

func TestChainedContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenant(ctx, "tenant-123")
	ctx = WithCluster(ctx, "cluster-001")
	ctx = WithUser(ctx, "user-456", "John Doe")

	tenantID, clusterID := ExtractAll(ctx)
	userID, userName, ok := UserFromContext(ctx)

	if tenantID != "tenant-123" {
		t.Errorf("Expected tenant-123, got %s", tenantID)
	}
	if clusterID != "cluster-001" {
		t.Errorf("Expected cluster-001, got %s", clusterID)
	}
	if !ok {
		t.Error("Expected user info to exist")
	}
	if userID != "user-456" || userName != "John Doe" {
		t.Errorf("Expected user-456/John Doe, got %s/%s", userID, userName)
	}
}