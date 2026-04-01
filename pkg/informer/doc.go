// Package informer provides Kubernetes SharedInformer management.
//
// This package is part of k8s-kit - a pure technical foundation library
// without business logic. It focuses on:
//
//   - SharedInformer creation and lifecycle management
//   - Informer pooling and reuse
//   - Resource type registration
//   - Cache synchronization
//
// This package does NOT handle:
//   - Which clusters to watch
//   - Tenant isolation
//   - Business-specific filtering
//
// Usage:
//
//	factory := informer.NewFactory()
//	defer factory.StopAll()
//
//	opts := informer.Options{
//	    Namespace:    "default",
//	    ResyncPeriod: 30 * time.Second,
//	    Lifecycle:    informer.LifecyclePersistent,
//	}
//
//	entry, err := factory.Create(client, opts)
//	if err != nil {
//	    return err
//	}
//
//	// Use entry.Factory.Core().V1().Pods().Lister()
//
package informer
