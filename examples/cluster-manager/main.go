// ClusterManager Example
//
// This example demonstrates how to use the ClusterManager to manage multiple Kubernetes clusters.
package main

import (
	"log"
	"os"
	"time"

	"github.com/guilinonline/k8s-kit/pkg/client"
	"github.com/guilinonline/k8s-kit/pkg/cluster"
)

func main() {
	//ctx := context.Background()

	// Create client factory
	clientFactory := client.NewFactory(
		client.WithTimeout(30*time.Second),
		client.WithQPS(100),
	)

	// Create ClusterManager with default health check config
	manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)

	// Set event callbacks
	manager.SetEventCallbacks(cluster.EventCallbacks{
		OnHealthy: func(id string) {
			log.Printf("Cluster %s is healthy", id)
		},
		OnUnhealthy: func(id string) {
			log.Printf("Cluster %s is unhealthy", id)
		},
		OnReconnected: func(id string) {
			log.Printf("Cluster %s reconnected", id)
		},
		OnInformerRecreate: func(id string) {
			log.Printf("Cluster %s Informer recreated", id)
		},
	})

	// Read kubeconfig (in production, this comes from your config source)
	kubeconfig, err := os.ReadFile("C:\\Users\\chenguilin\\code\\cglk8s-kit\\docs\\config-mock-1")
	if err != nil {
		panic(err)
	}

	// Register a cluster
	if err := manager.Register("cluster-001", kubeconfig,
		cluster.WithTenantID("tenant-001")); err != nil {
		log.Fatalf("Failed to register cluster: %v", err)
	}

	// Get a client for the cluster
	cli, err := manager.GetClient("cluster-001")
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	// Use the client
	_ = cli

	// List all clusters
	clusters := manager.List()
	log.Printf("Registered clusters: %v", clusters)

	// Get health status
	status, err := manager.GetHealthStatus("cluster-001")
	if err != nil {
		log.Fatalf("Failed to get health status: %v", err)
	}
	log.Printf("Cluster health: %s", status)

	// Clean up
	manager.Stop()
}
