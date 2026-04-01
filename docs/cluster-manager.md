# ClusterManager API Reference

## Overview

ClusterManager provides multi-cluster Kubernetes management, including cluster lifecycle management, health checking, auto-reconnection, and dynamic configuration refresh.

## Quick Start

```go
import (
    "github.com/guilinonline/k8s-kit/pkg/cluster"
    "github.com/guilinonline/k8s-kit/pkg/client"
)

// Create client factory
clientFactory := client.NewFactory()

// Create ClusterManager with default config
manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)

// Register a cluster
kubeconfig := []byte(`...`)
manager.Register("cluster-001", kubeconfig)

// Get client
cli, _ := manager.GetClient("cluster-001")
pods, _ := cli.Clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
```

## Core API

### NewManager

```go
func NewManager(factory *client.Factory, healthCfg HealthCheckConfig) *Manager
```

Creates a new ClusterManager.

### Register

```go
func (m *Manager) Register(id string, kubeconfig []byte, opts ...RegisterOption) error
```

Registers a new cluster.

Options:
- `WithTenantID(string)` - Set tenant ID for the cluster

### Unregister

```go
func (m *Manager) Unregister(id string) error
```

Unregisters a cluster and releases resources.

### GetClient

```go
func (m *Manager) GetClient(id string) (*client.ClusterClient, error)
```

Gets the client for a registered cluster.

### GetClientFromContext

```go
func (m *Manager) GetClientFromContext(ctx context.Context) (*client.ClusterClient, error)
```

Gets client using tenant and cluster info from context.

### GetHealthStatus

```go
func (m *Manager) GetHealthStatus(id string) (HealthStatus, error)
```

Gets the health status of a cluster.

### Stop

```go
func (m *Manager) Stop()
```

Stops the manager and all health check loops.

## Health Checking

The ClusterManager automatically performs health checks on registered clusters:

- **Interval**: Configurable (default: 30s)
- **Failure Threshold**: Configurable (default: 3 failures)
- **Auto Reconnect**: Enabled by default with exponential backoff

### Event Callbacks

```go
manager.SetEventCallbacks(cluster.EventCallbacks{
    OnHealthy:     func(id string) { ... },
    OnUnhealthy:   func(id string) { ... },
    OnReconnected: func(id string) { ... },
})
```

## Configuration Refresh (Hybrid Push + Pull)

The ClusterManager supports dynamic cluster configuration with a hybrid approach:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Hybrid Config Refresh                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────┐      Push (Real-time)      ┌─────────────┐  │
│   │ Config      │ ──────────────────────────▶ │ Manager     │  │
│   │ Source      │      Watch()                │ watchConfig │  │
│   │ (DB/API)    │                            │ Changes()   │  │
│   └─────────────┘                            └──────┬──────┘  │
│          │                                            │        │
│          │         Pull (Fallback)                    │        │
│          │    ┌──────────────────────┐               │        │
│          └───▶│  syncLoop()          │◀──────────────┘        │
│               │  (every 5 minutes)   │                        │
│               └──────────────────────┘                        │
│                                                                 │
│   ✓ Push: Immediate response to config changes                │
│   ✓ Pull: Fallback if watch connection fails                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Quick Start with Dynamic Config

```go
import (
    "context"
    "log"

    "github.com/guilinonline/k8s-kit/pkg/cluster"
    "github.com/guilinonline/k8s-kit/pkg/client"
)

// Step 1: Define your config source (e.g., from database)
type DBConfigProvider struct {
    // Your DB connection or API client
}

func (p *DBConfigProvider) GetAll(ctx context.Context) ([]cluster.ClusterConfig, error) {
    // Query clusters from your database
    // Return format:
    // return []cluster.ClusterConfig{
    //     {ID: "cluster-001", Kubeconfig: []byte(...), TenantID: "tenant-001"},
    //     {ID: "cluster-002", Kubeconfig: []byte(...), TenantID: "tenant-002"},
    // }, nil
    return []cluster.ClusterConfig{}, nil // TODO: implement
}

// Step 2: Implement Watch for Push mode (optional but recommended)
func (p *DBConfigProvider) Watch(ctx context.Context) (<-chan cluster.ClusterConfigChange, error) {
    ch := make(chan cluster.ClusterConfigChange, 10)

    // Your implementation depends on the data source:
    // - Database: Use NOTIFY/LISTEN or polling
    // - etcd: Watch API
    // - HTTP: Server-Sent Events
    //
    // Example (polling fallback):
    // go func() {
    //     for {
    //         select {
    //         case <-ctx.Done():
    //             close(ch)
    //             return
    //         case <-time.After(30 * time.Second):
    //             // Check for changes and send to channel
    //             if hasChanges {
    //                 ch <- cluster.ClusterConfigChange{
    //                     Type:       cluster.ChangeTypeAdd, // or Update, Delete
    //                     ClusterID:  "new-cluster",
    //                     Kubeconfig: []byte(...),
    //                     TenantID:   "tenant-001",
    //                 }
    //             }
    //         }
    //     }
    // }()

    return ch, nil
}

// Step 3: Use with ClusterManager
func main() {
    ctx := context.Background()

    clientFactory := client.NewFactory()
    manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)

    // Start with dynamic config (hybrid Push + Pull)
    configProvider := &DBConfigProvider{}
    if err := manager.Start(ctx, configProvider); err != nil {
        log.Fatalf("Failed to start manager: %v", err)
    }

    // Clusters are automatically loaded from DB on startup
    // and kept in sync via Push (Watch) + Pull (5min interval)

    // Your application logic here...
    // cli, _ := manager.GetClient("cluster-001")

    // Clean up
    manager.Stop()
}
```

### Interface Definitions

```go
// ClusterConfig represents a cluster's configuration
type ClusterConfig struct {
    ID        string
    Kubeconfig []byte
    TenantID  string
}

// ClusterConfigChange represents a change event
type ClusterConfigChange struct {
    Type      ChangeType  // ChangeTypeAdd, ChangeTypeUpdate, ChangeTypeDelete
    ClusterID string
    Kubeconfig []byte
    TenantID  string
}

// ConfigProvider is required for Start()
type ConfigProvider interface {
    GetAll(ctx context.Context) ([]ClusterConfig, error)
}

// ConfigWatcher is optional - enables Push mode
type ConfigWatcher interface {
    Watch(ctx context.Context) (<-chan ClusterConfigChange, error)
}
```

### Behavior

| Scenario | Behavior |
|----------|----------|
| Only implement `GetAll` | Pull only (sync every 5 minutes) |
| Implement `GetAll` + `Watch` | Hybrid: Push (real-time) + Pull (fallback) |
| Watch connection fails | Automatically falls back to Pull mode |
| Manager.Start() not called | Manual Register/Unregister only |