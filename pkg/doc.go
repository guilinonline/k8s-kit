// Package k8skit provides a pure technical foundation library for Kubernetes operations.
//
// k8s-kit is designed to be:
//   - Pure technical: No business logic, no tenant concepts, no configuration source assumptions
//   - DI-friendly: Easy to integrate with dependency injection containers
//   - Modular: Use only what you need
//   - Production-ready: Battle-tested patterns
//
// Subpackages:
//
//   - pkg/client: Kubernetes client initialization and management
//   - pkg/informer: SharedInformer management and lifecycle
//   - pkg/resource: Resource CRUD operations
//   - pkg/getter: Cache-based resource retrieval
//
// Example usage:
//
//	// Create client factory
//	clientFactory := client.NewFactory()
//
//	// Create client from kubeconfig (loaded from your config source)
//	cli, err := clientFactory.CreateFromKubeconfig(kubeconfigBytes)
//	if err != nil {
//	    panic(err)
//	}
//
//	// Create resource operator
//	operator := resource.NewOperator()
//
//	// Create a ConfigMap
//	cm := &corev1.ConfigMap{...}
//	err = operator.Create(ctx, cli, cm)
//
// For more examples, see the examples/ directory.
package k8skit
