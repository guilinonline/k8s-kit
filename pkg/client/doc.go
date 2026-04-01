// Package client provides Kubernetes client initialization and management.
//
// This package is part of k8s-kit - a pure technical foundation library
// without business logic. It focuses on:
//
//   - Client initialization from various sources (kubeconfig, REST config, in-cluster)
//   - Client configuration (timeout, QPS, burst, etc.)
//   - Client lifecycle management
//
// This package does NOT handle:
//   - Where configuration comes from (DB, file, etc.)
//   - Multi-tenancy concepts
//   - Business logic
//
// Usage:
//
//	factory := client.NewFactory()
//	cli, err := factory.CreateFromKubeconfig(kubeconfigBytes)
//	if err != nil {
//	    return err
//	}
//
//	// Use cli.Clientset or cli.RuntimeClient
//
package client
