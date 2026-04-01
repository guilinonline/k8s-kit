// Package cluster provides multi-cluster Kubernetes management, including cluster lifecycle,
// health checking, auto-reconnection, and dynamic configuration refresh.
//
// This package is part of k8s-kit - a pure technical foundation library
// without business logic. It focuses on:
//
//   - Cluster registration and unregistration
//   - Periodic health checking with failure thresholds
//   - Automatic reconnection with exponential backoff
//   - Dynamic configuration refresh (Push + Pull hybrid mode)
//   - Informer management per cluster (lazy creation)
//   - Event callbacks for cluster state changes
//
// This package does NOT handle:
//   - Where configuration comes from (DB, file, etc.)
//   - Business-level tenant isolation
//   - RBAC or authorization
//
// Example usage:
//
//	// Create manager
//	manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)
//
//	// Set event callbacks (optional)
//	manager.SetEventCallbacks(cluster.EventCallbacks{
//	    OnHealthy:   func(id string) { log.Printf("Cluster %s healthy", id) },
//	    OnUnhealthy: func(id string) { log.Printf("Cluster %s unhealthy", id) },
//	})
//
//	// Register a cluster
//	if err := manager.Register("cluster-001", kubeconfigBytes, cluster.WithTenantID("tenant-123")); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get client
//	cli, err := manager.GetClient("cluster-001")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use the client
//	pods, err := cli.Clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
//
// For more examples, see the examples/ directory.
package cluster
