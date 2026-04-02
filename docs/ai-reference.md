# k8s-kit AI 参考文档

本文档为 AI Agent 提供一站式 API 参考。请勿将此文档用于用户文档目的。

## 包结构概览

| 包 | 用途 |
|---|---|
| `pkg/client` | K8s 客户端创建工厂 |
| `pkg/cluster` | 多集群生命周期管理、健康检查、自动重连 |
| `pkg/pod` | Pod 日志获取、命令执行 |
| `pkg/tenant` | 多租户上下文传递 |
| `pkg/resource` | 资源 CRUD 操作 |
| `pkg/getter` | 基于 Informer 的缓存读取 |
| `pkg/informer` | SharedInformer 管理 |

---

## 完整使用示例

### 1. 创建客户端

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/guilinonline/k8s-kit/pkg/client"
)

func main() {
    // 读取 kubeconfig（实际项目中可能从配置中心获取）
    kubeconfig, err := os.ReadFile(os.ExpandEnv("$HOME/.kube/config"))
    if err != nil {
        panic(fmt.Sprintf("读取 kubeconfig 失败: %v", err))
    }

    // 创建工厂（使用推荐的默认 QPS/Burst）
    factory := client.NewFactory(
        client.WithTimeout(30 * time.Second),
        client.WithQPS(50),   // 默认 50，小集群可适当提高
        client.WithBurst(100), // 默认 100
    )

    // 创建客户端
    cli, err := factory.CreateFromKubeconfig(kubeconfig)
    if err != nil {
        panic(fmt.Sprintf("创建客户端失败: %v", err))
    }

    // 验证连接
    version, err := cli.Clientset.Discovery().ServerVersion()
    if err != nil {
        panic(fmt.Sprintf("连接集群失败: %v", err))
    }
    fmt.Printf("连接成功，K8s 版本: %s\n", version.GitVersion)

    _ = cli
}
```

### 2. 多集群管理

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/guilinonline/k8s-kit/pkg/client"
    "github.com/guilinonline/k8s-kit/pkg/cluster"
)

func main() {
    ctx := context.Background()

    // 创建工厂和管理器
    clientFactory := client.NewFactory()
    manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)
    defer manager.Stop() // 确保程序退出时清理资源

    // 设置事件回调（可选）
    manager.SetEventCallbacks(cluster.EventCallbacks{
        OnHealthy: func(id string) {
            fmt.Printf("集群 %s 恢复健康\n", id)
        },
        OnUnhealthy: func(id string) {
            fmt.Printf("集群 %s 不健康\n", id)
        },
        OnReconnected: func(id string) {
            fmt.Printf("集群 %s 重新连接成功\n", id)
        },
    })

    // 读取并注册集群
    kubeconfig1, _ := os.ReadFile("/path/to/cluster1/config")
    kubeconfig2, _ := os.ReadFile("/path/to/cluster2/config")

    if err := manager.Register("cluster-001", kubeconfig1, 
        cluster.WithTenantID("tenant-001")); err != nil {
        panic(fmt.Sprintf("注册集群1失败: %v", err))
    }

    if err := manager.Register("cluster-002", kubeconfig2,
        cluster.WithTenantID("tenant-002")); err != nil {
        panic(fmt.Sprintf("注册集群2失败: %v", err))
    }

    // 获取客户端
    cli1, err := manager.GetClient("cluster-001")
    if err != nil {
        panic(fmt.Sprintf("获取集群1客户端失败: %v", err))
    }

    // 使用客户端
    version, _ := cli1.Clientset.Discovery().ServerVersion()
    fmt.Printf("集群1版本: %s\n", version.GitVersion)

    // 查看所有集群
    clusters := manager.List()
    fmt.Printf("已注册集群: %v\n", clusters)

    // 查看健康状态
    status, _ := manager.GetHealthStatus("cluster-001")
    fmt.Printf("集群1健康状态: %s\n", status)
}
```

### 3. 多集群 + 多租户 + 动态配置（生产环境推荐）

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/guilinonline/k8s-kit/pkg/client"
    "github.com/guilinonline/k8s-kit/pkg/cluster"
    "github.com/guilinonline/k8s-kit/pkg/tenant"
    _ "github.com/go-sql-driver/mysql"
)

// DBConfigProvider 从数据库获取集群配置
type DBConfigProvider struct {
    db       *sql.DB
    changeCh chan cluster.ClusterConfigChange
}

// GetAll 查询所有集群配置
func (p *DBConfigProvider) GetAll(ctx context.Context) ([]cluster.ClusterConfig, error) {
    rows, err := p.db.QueryContext(ctx, "SELECT id, kubeconfig, tenant_id FROM clusters WHERE enabled = 1")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var configs []cluster.ClusterConfig
    for rows.Next() {
        var cfg cluster.ClusterConfig
        if err := rows.Scan(&cfg.ID, &cfg.Kubeconfig, &cfg.TenantID); err != nil {
            continue
        }
        configs = append(configs, cfg)
    }
    return configs, nil
}

// Watch 实现 Push 模式（可选）
func (p *DBConfigProvider) Watch(ctx context.Context) (<-chan cluster.ClusterConfigChange, error) {
    p.changeCh = make(chan cluster.ClusterConfigChange, 100)

    // 示例：轮询数据库检查变化
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                close(p.changeCh)
                return
            case <-ticker.C:
                // 实际项目中：对比上次查询结果，检测变化
                // 这里简化示例
            }
        }
    }()

    return p.changeCh, nil
}

// NotifyChange 业务层主动通知配置变化
func (p *DBConfigProvider) NotifyChange(change cluster.ClusterConfigChange) {
    if p.changeCh != nil {
        p.changeCh <- change
    }
}

func main() {
    ctx := context.Background()

    // 初始化数据库连接
    db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/k8s")
    if err != nil {
        panic(err)
    }

    // 创建配置提供者
    provider := &DBConfigProvider{db: db}

    // 创建管理器
    clientFactory := client.NewFactory()
    manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)
    defer manager.Stop()

    // 启动动态配置（混合模式：Push + Pull）
    if err := manager.Start(ctx, provider); err != nil {
        panic(fmt.Sprintf("启动管理器失败: %v", err))
    }

    // ===== 使用示例 =====

    // 场景1：普通使用（指定集群）
    cli, err := manager.GetClient("cluster-001")
    if err != nil {
        panic(err)
    }
    _ = cli

    // 场景2：多租户场景（从 HTTP 请求中提取租户信息）
    // 假设从 JWT 或 Header 中提取到租户和集群信息
    tenantID := "tenant-001"
    clusterID := "cluster-001"

    // 设置上下文
    ctx = tenant.WithTenantAndCluster(ctx, tenantID, clusterID)

    // 使用上下文获取客户端（自动根据租户和集群路由）
    cli2, err := manager.GetClientFromContext(ctx)
    if err != nil {
        panic(fmt.Sprintf("获取客户端失败: %v", err))
    }

    version, _ := cli2.Clientset.Discovery().ServerVersion()
    fmt.Printf("租户 %s 的集群 %s 版本: %s\n", tenantID, clusterID, version)

    // 场景3：业务层修改配置后主动通知
    // 比如在管理后台更新了集群配置
    provider.NotifyChange(cluster.ClusterConfigChange{
        Type:       cluster.ChangeTypeUpdate,
        ClusterID:  "cluster-001",
        Kubeconfig: []byte("new kubeconfig..."),
        TenantID:   "tenant-001",
    })
}
```

### 4. 资源操作（完整 CRUD + 缓存）

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
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/apimachinery/pkg/types"
)

func main() {
    ctx := context.Background()

    // 创建客户端
    factory := client.NewFactory()
    kubeconfig, _ := os.ReadFile(os.ExpandEnv("$HOME/.kube/config"))
    cli, err := factory.CreateFromKubeconfig(kubeconfig)
    if err != nil {
        panic(err)
    }

    // ===== 推荐方式：使用带缓存的 Operator =====
    operator := resource.NewOperatorWithClient(cli)
    defer operator.Stop()

    // ----- Create 创建资源 -----
    newPod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
            Labels: map[string]string{
                "app": "test",
            },
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:  "nginx",
                    Image: "nginx:latest",
                },
            },
        },
    }

    if err := operator.Create(ctx, cli, newPod); err != nil {
        panic(fmt.Sprintf("创建 Pod 失败: %v", err))
    }
    fmt.Println("Pod 创建成功")

    // ----- Get 获取资源（从缓存读）-----
    pod := &corev1.Pod{}
    if err := operator.Get(ctx, cli, pod, types.NamespacedName{
        Namespace: "default",
        Name:      "test-pod",
    }); err != nil {
        panic(fmt.Sprintf("获取 Pod 失败: %v", err))
    }
    fmt.Printf("获取 Pod: %s, 状态: %s\n", pod.Name, pod.Status.Phase)

    // ----- List 列表查询（从缓存读）-----
    podList := &corev1.PodList{}
    if err := operator.List(ctx, cli, podList,
        resource.WithNamespace("default"),
        resource.WithLabelSelector(labels.SelectorFromSet(labels.Set{"app": "test"})),
    ); err != nil {
        panic(fmt.Sprintf("列表查询失败: %v", err))
    }
    fmt.Printf("查询到 %d 个 Pod\n", len(podList.Items))

    // ----- List 后客户端过滤 + 分页（适合数据量较小的场景）-----
    // 场景1：先 LabelSelector 减少数据量，再内存过滤和分页
    operator.List(ctx, cli, podList,
        resource.WithNamespace("default"),
        resource.WithLabelSelector(labels.SelectorFromSet(labels.Set{"app": "nginx"})),
    )
    filtered := resource.FilterByName(podList.Items, "web-")  // 再从结果中过滤
    fmt.Printf("Label=nginx 且名称包含 'web-' 的 Pod: %d 个\n", len(filtered))

    // 场景2：查全量 → 过滤 → 内存分页（适合导出、统计类场景）
    operator.List(ctx, cli, podList)  // 不传 Limit，查全量
    filtered = resource.FilterByName(podList.Items, "nginx")
    pageNum := 1
    pageSize := 10
    pageData, hasMore := resource.Paginate(filtered, pageSize, (pageNum-1)*pageSize)
    fmt.Printf("第 %d 页 %d 条，总共 %d 条，是否还有更多: %v\n",
        pageNum, len(pageData), len(filtered), hasMore)

    // 其他匹配方式示例
    webPods := resource.FilterByName(podList.Items, "web-", resource.MatchPrefix)
    prodPods := resource.FilterByName(podList.Items, "-prod", resource.MatchSuffix)

    // ----- Update 更新资源 -----
    pod.Labels["version"] = "v2"
    if err := operator.Update(ctx, cli, pod); err != nil {
        panic(fmt.Sprintf("更新 Pod 失败: %v", err))
    }
    fmt.Println("Pod 更新成功")

    // ----- Delete 删除资源 -----
    if err := operator.Delete(ctx, cli, pod); err != nil {
        panic(fmt.Sprintf("删除 Pod 失败: %v", err))
    }
    fmt.Println("Pod 删除成功")

    // ===== 无缓存模式（小规模/低频场景）=====
    // operator := resource.NewOperator()
    // 所有操作直接调用 APIServer，不走缓存
}
```

### 5. Pod 日志（简单 + 流式 + 回调）

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "os"
    "time"

    "github.com/guilinonline/k8s-kit/pkg/client"
    "github.com/guilinonline/k8s-kit/pkg/pod"
)

func main() {
    ctx := context.Background()

    // 创建客户端
    factory := client.NewFactory()
    kubeconfig, _ := os.ReadFile(os.ExpandEnv("$HOME/.kube/config"))
    cli, _ := factory.CreateFromKubeconfig(kubeconfig)

    operator := pod.NewOperator()

    namespace := "default"
    podName := "nginx-pod"

    // ----- 方式1：简单获取（适合小日志）-----
    logs, err := operator.GetLogsSimple(ctx, cli, namespace, podName,
        pod.WithContainer("nginx"),
        pod.WithTailLines(100),
        pod.WithTimestamps(true),
    )
    if err != nil {
        // 判断错误类型
        if pod.IsPodNotFound(err) {
            fmt.Println("Pod 不存在")
        } else if pod.IsContainerNotFound(err) {
            fmt.Println("容器不存在")
        } else {
            panic(fmt.Sprintf("获取日志失败: %v", err))
        }
    }
    fmt.Printf("日志内容:\n%s\n", logs)

    // ----- 方式2：流式获取（适合大日志）-----
    stream, err := operator.GetLogsStream(ctx, cli, namespace, podName,
        pod.WithTailLines(1000),
    )
    if err != nil {
        panic(err)
    }
    defer stream.Close()

    // 逐行读取
    scanner := bufio.NewScanner(stream)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Println(line)
    }

    // ----- 方式3：回调方式（实时监控）-----
    fmt.Println("开始实时监控日志（按 Ctrl+C 停止）...")
    err = operator.TailLogs(ctx, cli, namespace, podName,
        func(line string) {
            fmt.Printf("[LOG] %s\n", line)
        },
        pod.WithTailLines(10),
        pod.WithContainer("nginx"),
    )
    if err != nil {
        panic(fmt.Sprintf("日志监控失败: %v", err))
    }
}
```

### 6. Pod Exec（简单 + 流式交互）

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/guilinonline/k8s-kit/pkg/client"
    "github.com/guilinonline/k8s-kit/pkg/pod"
)

func main() {
    ctx := context.Background()

    // 创建客户端
    factory := client.NewFactory()
    kubeconfig, _ := os.ReadFile(os.ExpandEnv("$HOME/.kube/config"))
    cli, _ := factory.CreateFromKubeconfig(kubeconfig)

    operator := pod.NewOperator()

    namespace := "default"
    podName := "nginx-pod"

    // ----- 方式1：简单执行（适合一次性命令）-----
    result, err := operator.ExecSimple(ctx, cli, namespace, podName,
        []string{"ls", "-la", "/"},
        pod.WithExecContainer("nginx"),
        pod.WithExecTimeout(10*time.Second),
    )
    if err != nil {
        // 判断错误类型
        if pod.IsPodNotFound(err) {
            fmt.Println("Pod 不存在")
        } else if pod.IsContainerNotFound(err) {
            fmt.Println("容器不存在")
        } else if pod.IsTimeout(err) {
            fmt.Println("执行超时")
        } else {
            panic(fmt.Sprintf("执行失败: %v", err))
        }
    }

    fmt.Printf("ExitCode: %d\n", result.ExitCode)
    fmt.Printf("Stdout:\n%s\n", result.Stdout)
    if result.Stderr != "" {
        fmt.Printf("Stderr:\n%s\n", result.Stderr)
    }

    // ----- 方式2：流式执行（适合交互式命令）-----
    session, err := operator.ExecStream(ctx, cli, namespace, podName,
        []string{"sh"},
        pod.WithExecContainer("nginx"),
    )
    if err != nil {
        panic(err)
    }
    defer session.Close()

    // 发送命令
    session.Write([]byte("echo Hello from Pod\n"))
    session.Write([]byte("ls -la\n"))
    session.Write([]byte("exit\n"))

    // 读取输出
    buf := make([]byte, 1024)
    for {
        n, err := session.Read(buf)
        if err != nil {
            break
        }
        fmt.Print(string(buf[:n]))
    }

    // 等待执行完成
    if err := session.Wait(); err != nil {
        fmt.Printf("执行结束，错误: %v\n", err)
    }
}
```

---

## API 速查表

### pkg/client

```go
import "github.com/guilinonline/k8s-kit/pkg/client"

// 创建工厂
factory := client.NewFactory(
    client.WithTimeout(30 * time.Second),
    client.WithQPS(50),
    client.WithBurst(100),
)

// 从 kubeconfig 创建客户端
cli, err := factory.CreateFromKubeconfig(kubeconfig []byte)

// 从集群内创建（用于 Pod 内运行）
cli, err := factory.CreateFromInCluster()

// ClusterClient 结构
type ClusterClient struct {
    Clientset    *kubernetes.Clientset      // client-go 原生客户端
    RuntimeClient runtimeclient.Client      // controller-runtime 客户端
    RESTConfig   *rest.Config               // REST 配置
}
```

### pkg/cluster

```go
import "github.com/guilinonline/k8s-kit/pkg/cluster"

// 创建管理器
manager := cluster.NewManager(factory, cluster.DefaultHealthCheckConfig)
defer manager.Stop()

// 注册集群
manager.Register(id, kubeconfig, cluster.WithTenantID("tenant-001"))

// 获取客户端
cli, err := manager.GetClient("cluster-001")
cli, err := manager.GetClientFromContext(ctx)  // 从上下文获取

// 启动动态配置
manager.Start(ctx, configProvider)

// 事件回调
manager.SetEventCallbacks(cluster.EventCallbacks{
    OnHealthy:     func(id string) { ... },
    OnUnhealthy:   func(id string) { ... },
    OnReconnected: func(id string) { ... },
})
```

### pkg/resource

```go
import "github.com/guilinonline/k8s-kit/pkg/resource"

// 带缓存模式（推荐）
operator := resource.NewOperatorWithClient(cli)
defer operator.Stop()

// 无缓存模式
operator := resource.NewOperator()

// CRUD 操作
operator.Get(ctx, cli, obj, key)
operator.List(ctx, cli, list, opts...)
operator.Create(ctx, cli, obj, opts...)
operator.Update(ctx, cli, obj, opts...)
operator.Patch(ctx, cli, obj, patch, opts...)
operator.Delete(ctx, cli, obj, opts...)

// List 选项
resource.WithNamespace(ns)
resource.WithLabelSelector(selector)
resource.WithFieldSelector(selector)
resource.WithLimit(n)          // 分页限制
resource.WithContinue(token)   // 分页 token

// 客户端过滤函数（List 后内存过滤）
//
// ⚠️⚠️⚠️ 重要限制（其他 AI 必读）：
//
// 1. 【问题】K8s APIServer 不支持按名称模糊查询，所以此函数在客户端内存中过滤
// 2. 【限制】如果先使用 WithLimit 分页，再调用此函数，可能出现"一页只有几条"的问题
// 3. 【适用】适合以下场景：
//      - 数据量较小（< 1000 条）
//      - 配合 LabelSelector 先减少数据量后再过滤
//      - 不分页展示（如导出、统计、批量操作）
// 4. 【不适用】不适合直接用于大数据量的分页列表查询
//
// 【推荐用法】配合 LabelSelector 先减少数据量：
//   operator.List(ctx, cli, podList, resource.WithLabelSelector(selector))
//   filtered := resource.FilterByName(podList.Items, "nginx")
//
// 【错误用法】直接在分页查询后过滤（会导致分页异常）：
//   operator.List(ctx, cli, podList, resource.WithLimit(10))  // 先分页
//   filtered := resource.FilterByName(podList.Items, "nginx") // ❌ 再过滤，可能只剩2条
//
// 匹配方式：
filtered := resource.FilterByName(items, "nginx")                       // 包含（默认）
filtered := resource.FilterByName(items, "web-", resource.MatchPrefix)  // 前缀
filtered := resource.FilterByName(items, "-prod", resource.MatchSuffix) // 后缀
filtered := resource.FilterByName(items, "^web-.*-prod$", resource.MatchRegex) // 正则

// 内存分页（配合过滤使用）
// ⚠️ 注意：此函数在客户端内存中分页，适合数据量较小的场景
pageData, hasMore := resource.Paginate(filtered, limit, offset)

// 完整示例：查全量 → 过滤 → 内存分页
operator.List(ctx, cli, podList)  // 不传 Limit，查全量
filtered := resource.FilterByName(podList.Items, "nginx")
pageData, hasMore := resource.Paginate(filtered, 10, (pageNum-1)*10)  // 第 pageNum 页，每页 10 条
fmt.Printf("本页 %d 条，是否还有更多: %v\n", len(pageData), hasMore)
```

### pkg/pod

```go
import "github.com/guilinonline/k8s-kit/pkg/pod"

operator := pod.NewOperator()

// 日志
logs, err := operator.GetLogsSimple(ctx, cli, ns, name, opts...)
stream, err := operator.GetLogsStream(ctx, cli, ns, name, opts...)
err := operator.TailLogs(ctx, cli, ns, name, handleLine, opts...)

// Exec
result, err := operator.ExecSimple(ctx, cli, ns, name, command, opts...)
session, err := operator.ExecStream(ctx, cli, ns, name, command, opts...)

// 日志选项
pod.WithContainer(name)
pod.WithTailLines(n)
pod.WithTimestamps(bool)
pod.WithSinceTime(time)
pod.WithLimitBytes(n)

// Exec 选项
pod.WithExecContainer(name)
pod.WithExecTimeout(d)

// 错误判断
pod.IsPodNotFound(err)
pod.IsContainerNotFound(err)
pod.IsTimeout(err)
pod.IsForbidden(err)
pod.IsConnectionLost(err)
```

### pkg/tenant

```go
import "github.com/guilinonline/k8s-kit/pkg/tenant"

// 设置
ctx = tenant.WithTenant(ctx, "tenant-001")
ctx = tenant.WithCluster(ctx, "cluster-001")
ctx = tenant.WithTenantAndCluster(ctx, "tenant-001", "cluster-001")

// 获取
tenantID := tenant.FromContext(ctx)
clusterID := tenant.ClusterFromContext(ctx)
tenantID, clusterID = tenant.ExtractAll(ctx)
```

### pkg/getter（高级用法）

```go
import "github.com/guilinonline/k8s-kit/pkg/getter"

// 需要手动管理 Informer 时使用
cache := getter.NewCacheReader(factory)

pod, err := cache.GetPod(ns, name)
pods, err := cache.ListPods(ns, selector)
svcs, err := cache.ListServices(ns, selector)
deploys, err := cache.ListDeployments(ns, selector)
```

---

## 常见 import 汇总

```go
import (
    // kit 包
    "github.com/guilinonline/k8s-kit/pkg/client"
    "github.com/guilinonline/k8s-kit/pkg/cluster"
    "github.com/guilinonline/k8s-kit/pkg/pod"
    "github.com/guilinonline/k8s-kit/pkg/resource"
    "github.com/guilinonline/k8s-kit/pkg/tenant"
    "github.com/guilinonline/k8s-kit/pkg/getter"

    // K8s API
    corev1 "k8s.io/api/core/v1"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/apimachinery/pkg/fields"
)
```

---

## 设计原则

1. **纯技术** - 无业务逻辑，无租户概念，无配置来源假设
2. **依赖注入友好** - 所有组件可轻松 mock
3. **按需使用** - 只导入需要的包
4. **显式生命周期** - 由使用者控制创建和销毁

---

## 注意事项

1. **记得调用 Stop()** - Operator 和 Manager 都需要在程序退出时调用 Stop() 清理资源
2. **错误处理** - 使用 pkg/pod 提供的错误判断函数来区分错误类型
3. **QPS 设置** - 小集群可以适当提高 QPS，大集群保持默认值
4. **缓存模式** - 高频读场景用 `NewOperatorWithClient`，低频场景用 `NewOperator`
