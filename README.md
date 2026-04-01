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
- `ClientOptions`: Configuration (timeout, QPS, etc.)

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

## License

MIT License - see LICENSE file for details.
