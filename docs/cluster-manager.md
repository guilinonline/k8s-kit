# ClusterManager API Reference

## Overview

ClusterManager provides multi-cluster Kubernetes management, including cluster lifecycle management, health checking, auto-reconnection, and dynamic configuration refresh.

## Quick Start

```go
import (
    "github.com/seaman/k8s-kit/pkg/cluster"
    "github.com/seaman/k8s-kit/pkg/client"
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

## Configuration Refresh

Start with dynamic configuration:

```go
manager.Start(ctx, configProvider)
```

Where `ConfigProvider` implements:
- `GetAll(ctx) ([]ClusterConfig, error)`
- `Watch(ctx) (<-chan ClusterConfigChange, error)` (optional)