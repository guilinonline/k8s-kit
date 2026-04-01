# Multi-Tenancy Guide

## Overview

k8s-kit provides multi-tenancy support through Context-based tenant information passing. The KIT does not implement business-level isolation but provides the infrastructure for tenant identification.

## Core Concepts

### Context-Based Tenant Passing

```go
import "github.com/seaman/k8s-kit/pkg/tenant"

// Inject tenant
ctx := tenant.WithTenant(context.Background(), "tenant-123")

// Extract tenant
tenantID := tenant.FromContext(ctx) // "tenant-123"

// Default tenant if not set
ctx := context.Background()
tenantID := tenant.FromContext(ctx) // "default"
```

### Cluster Context

```go
// Set both tenant and cluster
ctx := tenant.WithTenantAndCluster(ctx, "tenant-123", "cluster-001")

// Get cluster from context
clusterID := tenant.ClusterFromContext(ctx)
```

### With ClusterManager

```go
// Get client with tenant context
cli, err := manager.GetClientFromContext(ctx)
```

## Best Practices

1. **Always use tenant context** in multi-tenant scenarios
2. **Implement access control** in business layer
3. **Use middleware** to extract tenant from requests

## Example: HTTP Middleware

```go
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := extractTenantFromJWT(r) // Your implementation
        ctx := tenant.WithTenant(r.Context(), tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```