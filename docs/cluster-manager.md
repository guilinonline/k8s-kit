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

**核心概念**：业务层实现 `Watch()` 返回一个 channel，**业务主动向这个 channel 推送变更**，kit 被动监听。

```go
import (
    "context"
    "log"

    "github.com/guilinonline/k8s-kit/pkg/cluster"
    "github.com/guilinonline/k8s-kit/pkg/client"
)

// Step 1: Define your config source (e.g., from database)
type DBConfigProvider struct {
    db *sql.DB
    // 导出 channel，让业务层其他地方可以推送变更
    ChangeCh chan cluster.ClusterConfigChange
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

// Step 2: 实现 Watch - 返回一个 channel，kit 会监听这个 channel
//        业务层在发现配置变化时，主动向这个 channel 推送变更
func (p *DBConfigProvider) Watch(ctx context.Context) (<-chan cluster.ClusterConfigChange, error) {
    p.ChangeCh = make(chan cluster.ClusterConfigChange, 10)

    // 方式 A: 业务自己轮询数据库，发现变化后推送（示例）
    go func() {
        for {
            select {
            case <-ctx.Done():
                close(p.ChangeCh)
                return
            case <-time.After(30 * time.Second):
                // 扫描数据库变化，推送增量
                if changes := p.scanChangesFromDB(); len(changes) > 0 {
                    for _, c := range changes {
                        p.ChangeCh <- c
                    }
                }
            }
        }
    }()

    return p.ChangeCh, nil
}

// 业务层调用示例: 修改集群配置后，主动推送变更
func (s *Server) UpdateClusterAPI(clusterID string, kubeconfig []byte) error {
    // 1. 更新数据库
    s.db.UpdateCluster(clusterID, kubeconfig)

    // 2. 主动通知 k8s-kit（通过 channel）
    s.provider.ChangeCh <- cluster.ClusterConfigChange{
        Type:       cluster.ChangeTypeUpdate,
        ClusterID:  clusterID,
        Kubeconfig: kubeconfig,
    }
    return nil
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

| Scenario | Business Implementation Required |
|----------|----------------------------------|
| Only `GetAll` | **只需实现 `GetAll()`** → Pull 自动工作，每 5 分钟同步 |
| `GetAll` + `Watch` | **只需实现 `GetAll()` + `Watch()`** → Push + Pull 混合 |

**关键点**：`GetAll()` 是给 Pull 用的，数据源就是业务实现 `GetAll()` 时查询的地方（MySQL/API/文件）。不需要额外配置。
| Watch connection fails | Automatically falls back to Pull mode |
| Manager.Start() not called | Manual Register/Unregister only |