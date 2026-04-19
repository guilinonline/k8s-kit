# k8s-kit

A pure technical foundation library for Kubernetes operations without business logic.

## Overview

k8s-kit is a Go library that provides a clean, dependency-injection friendly abstraction over Kubernetes client-go and controller-runtime. It is designed to be:

- **Pure technical**: No business logic, no tenant concepts, no configuration source assumptions
- **DI-friendly**: All components accept interfaces, easy to mock and test
- **Modular**: Use only what you need - client, informer, or resource operator
- **Production-ready**: Battle-tested patterns from real-world K8s operations

## Installation

```bash
go get github.com/guilinonline/k8s-kit
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/guilinonline/k8s-kit/pkg/client"
    "github.com/guilinonline/k8s-kit/pkg/resource"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
    ctx := context.Background()

    // Read kubeconfig (in real app, this comes from your config source)
    kubeconfig, _ := os.ReadFile(os.ExpandEnv("$HOME/.kube/config"))

    // Create client
    factory := client.NewFactory()
    cli, err := factory.CreateFromKubeconfig(kubeconfig, client.WithTimeout(30))
    if err != nil {
        panic(err)
    }

    // Create resource operator
    operator := resource.NewOperator()

    // Create a ConfigMap
    cm := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "my-config",
            Namespace: "default",
        },
        Data: map[string]string{
            "key": "value",
        },
    }

    if err := operator.Create(ctx, cli, cm); err != nil {
        panic(err)
    }

    fmt.Println("ConfigMap created successfully!")
}
```

## Packages

### `pkg/client`

Client initialization and management.

- `Factory`: Creates ClusterClient from various sources
- `ClusterClient`: Holds all K8s clients (Clientset, RuntimeClient, RESTConfig)
- `ClientOptions`: Configuration (timeout, QPS, DialContext, etc.)

#### Custom Dial（隧道/代理）

通过 `WithDialContext` 注入自定义拨号函数，适用于内网集群通过代理/隧道访问的场景：

```go
factory := client.NewFactory()
cli, err := factory.CreateFromKubeconfig(kubeconfig,
    client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
        // 通过隧道连接内网集群
        return tunnelDial(ctx, network, addr)
    }),
)
```

### `pkg/cluster`

Multi-cluster management with health checking and auto-reconnection.

- `Manager`: Multi-cluster lifecycle management
- `ConfigProvider`: Interface for loading cluster configs from any source (DB, file, etc.)
- `ConfigWatcher`: Optional interface for real-time config change notifications
- `RegisterOptions`: Registration options (TenantID, DialContext)

#### 注册集群时注入隧道

```go
manager.Register(clusterID, kubeconfig,
    cluster.WithTenantID("tenant-1"),
    cluster.WithDialContext(dialFn),  // 可选，内网集群走隧道
)
```

#### ConfigProvider 携带 DialContext

实现 `ConfigProvider` 接口时，可为内网集群提供 `DialContext`：

```go
func (r *MyRepository) GetAll(ctx context.Context) ([]cluster.ClusterConfig, error) {
    // ...
    for i, c := range clusters {
        cfg := cluster.ClusterConfig{
            ID:         c.ID,
            Kubeconfig: c.Kubeconfig,
            TenantID:   c.TenantID,
        }
        if c.IsInternal && r.tunnelDial != nil {
            cfg.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
                return r.tunnelDial(ctx, c.TenantID, c.GroupID, network, addr)
            }
        }
        result[i] = cfg
    }
    return result, nil
}
```

Manager 在加载/注册/更新集群时自动将 `DialContext` 透传到 `rest.Config.Dial`，无需外部 hack。

### `pkg/informer`

SharedInformer management.

- `Factory`: Creates and manages SharedInformer instances
- `Entry`: Managed Informer with lifecycle tracking
- `Options`: Informer configuration (namespace, resync period, lifecycle)

### `pkg/resource`

Resource operations.

- `Operator`: CRUD operations on K8s resources
- `Getter`: Get and List operations with caching support
- Options for each operation type

## Design Principles

### 1. No Business Logic

This library knows nothing about:
- Where configuration comes from (DB, file, env, etc.)
- Multi-tenancy concepts
- Your business rules

**You** bring your own configuration source and tenant management.

### 2. Dependency Injection Friendly

All major components are interfaces or concrete structs that can be easily mocked:

```go
type MyService struct {
    clientFactory   *client.Factory
    informerFactory *informer.Factory
    resourceOp      *resource.Operator
}
```

### 3. Composable

Use only what you need:

- Just need a client? Use `pkg/client`
- Need caching? Add `pkg/informer`
- Need operations? Add `pkg/resource`

### 4. Explicit Lifecycle

No magic. You control when things are created and destroyed:

```go
factory := informer.NewFactory()
defer factory.StopAll() // You decide when to stop
```

## Examples

See `examples/` directory for complete examples:

- `basic/main.go`: Basic client and resource operations
- `informer/main.go`: Using SharedInformer for caching
- `advanced/main.go`: Advanced patterns and best practices

## Contributing

Contributions are welcome! Please:

1. Keep it pure technical - no business logic
2. Maintain DI-friendly design
3. Add tests for new features
4. Update documentation

## Author

**桂林 (C.Guilin)**

- GitHub: [@guilinonline](https://github.com/guilinonline)
- 博客: [https://gl.sh.cn](https://gl.sh.cn)
- 职业: 程序员 / 运维工程师
- 坐标: 中国 · 桂林
- 认证: **CKA** (Certified Kubernetes Administrator)、**CKS** (Certified Kubernetes Security Specialist)

具备大型互联网运维架构设计能力，曾任职于头部企业负责运维研发团队技术管理。专注于 Kubernetes、云原生技术和 Go 语言开发，热爱开源分享。

## License

MIT License - see LICENSE file for details.
