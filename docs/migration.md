# Migration Guide

## From client-go to k8s-kit

### Before

```go
// Create client directly
restConfig, _ := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
clientset, _ := kubernetes.NewForConfig(restConfig)

// Use client
pods, _ := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
```

### After

```go
// Create ClusterManager
manager := cluster.NewManager(clientFactory, config)

// Register cluster
manager.Register("cluster-001", kubeconfig)

// Get client through manager
cli, _ := manager.GetClient("cluster-001")

// Use client
pods, _ := cli.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
```

### Benefits

- Automatic client lifecycle management
- Health checking and auto-reconnection
- Multi-cluster support
- Informer caching

## From Single-Tenant to Multi-Tenant

### Before

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    cli, _ := manager.GetClient("cluster-001")
    pods, _ := cli.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
}
```

### After

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    tenantID := extractTenant(r)
    clusterID := extractCluster(r)

    ctx := tenant.WithTenantAndCluster(r.Context(), tenantID, clusterID)
    cli, _ := manager.GetClientFromContext(ctx)
    pods, _ := cli.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
}
```

## Breaking Changes

This release is **fully backward compatible**. All new features are additive.