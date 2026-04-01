// Package getter provides cache-based resource retrieval using SharedInformer.
//
// This package is part of k8s-kit and provides read operations that leverage
// the Informer cache for better performance compared to direct API server calls.
//
// Example:
//
//	// Create Informer first
//	entry, _ := informerFactory.Create(client, opts)
//
//	// Create cache reader
//	reader := getter.NewCacheReader(entry.Factory)
//
//	// List pods from cache
//	pods, _ := reader.ListPods("default", labels.Everything())
//
package getter
