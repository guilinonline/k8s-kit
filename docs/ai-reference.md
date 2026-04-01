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

## 快速示例

### 1. 创建客户端

```go
import "github.com/guilinonline/k8s-kit/pkg/client"

factory := client.NewFactory(
    client.WithTimeout(30 * time.Second),
    client.WithQPS(100),
)
cli, err := factory.CreateFromKubeconfig(kubeconfig)
```

### 2. 多集群管理

```go
import "github.com/guilinonline/k8s-kit/pkg/cluster"

manager := cluster.NewManager(clientFactory, cluster.DefaultHealthCheckConfig)
manager.Register("cluster-001", kubeconfig, cluster.WithTenantID("tenant-001"))
cli, _ := manager.GetClient("cluster-001")
manager.Stop()
```

### 3. Pod 日志

```go
import "github.com/guilinonline/k8s-kit/pkg/pod"

operator := pod.NewOperator()
logs, _ := operator.GetLogsSimple(ctx, cli, "default", "nginx-pod",
    pod.WithTailLines(100),
    pod.WithContainer("nginx"),
)
```

### 4. Pod Exec

```go
result, _ := operator.ExecSimple(ctx, cli, "default", "nginx-pod",
    []string{"ls", "-la"},
    pod.WithExecTimeout(10*time.Second),
)
fmt.Println(result.ExitCode, result.Stdout)
```

### 5. 多租户上下文

```go
import "github.com/guilinonline/k8s-kit/pkg/tenant"

// 设置租户和集群
ctx = tenant.WithTenantAndCluster(context.Background(), "tenant-123", "cluster-001")

// 从上下文获取客户端
cli, _ := manager.GetClientFromContext(ctx)

// 提取信息
tenantID, clusterID := tenant.ExtractAll(ctx)
```

### 6. 资源操作

```go
import "github.com/guilinonline/k8s-kit/pkg/resource"

operator := resource.NewOperator()
podList := &corev1.PodList{}
operator.List(ctx, cli, podList,
    resource.WithNamespace("default"),
    resource.WithLimit(10),
)
```

---

## API 参考

### pkg/client

```go
// Factory 创建客户端
func NewFactory(opts ...FactoryOption) *Factory
func (f *Factory) CreateFromKubeconfig(kubeconfig []byte, opts ...Option) (*ClusterClient, error)

// ClusterClient 包含所有 K8s 客户端
type ClusterClient struct {
    Clientset    *kubernetes.Clientset
    RuntimeClient runtimeclient.Client
    RESTConfig   *rest.Config
}
```

### pkg/cluster

```go
// Manager 多集群管理器
func NewManager(factory *client.Factory, healthCfg HealthCheckConfig) *Manager

// 注册/注销集群
func (m *Manager) Register(id string, kubeconfig []byte, opts ...RegisterOption) error
func (m *Manager) Unregister(id string) error

// 获取客户端
func (m *Manager) GetClient(id string) (*client.ClusterClient, error)
func (m *Manager) GetClientFromContext(ctx context.Context) (*client.ClusterClient, error)

// 健康检查
func (m *Manager) GetHealthStatus(id string) (HealthStatus, error)

// 列表
func (m *Manager) List() []string

// 生命周期
func (m *Manager) Stop()
func (m *Manager) Start(ctx context.Context, provider ConfigProvider) error

// 注册选项
func WithTenantID(tenantID string) RegisterOption

// 配置提供者接口
type ConfigProvider interface {
    GetAll(ctx context.Context) ([]ClusterConfig, error)
    Watch(ctx context.Context) (<-chan ClusterConfigChange, error)
}
```

### pkg/pod

```go
type Operator struct{}

func NewOperator() *Operator

// 日志获取
func (o *Operator) GetLogsSimple(ctx context.Context, cli *client.ClusterClient, namespace, podName string, opts ...LogOption) (string, error)
func (o *Operator) GetLogsStream(ctx context.Context, cli *client.ClusterClient, namespace, podName string, opts ...LogOption) (io.ReadCloser, error)
func (o *Operator) TailLogs(ctx context.Context, cli *client.ClusterClient, namespace, podName string, handleLine func(line string), opts ...LogOption) error

// 命令执行
func (o *Operator) ExecSimple(ctx context.Context, cli *client.ClusterClient, namespace, podName string, command []string, opts ...ExecOption) (*ExecResult, error)
func (o *Operator) ExecStream(ctx context.Context, cli *client.ClusterClient, namespace, podName string, command []string, opts ...ExecOption) (*ExecSession, error)

// 日志选项
func WithContainer(name string) LogOption
func WithTailLines(n int64) LogOption
func WithTimestamps(bool) LogOption
func WithSinceTime(time.Time) LogOption
func WithLimitBytes(n int) LogOption

// Exec 选项
func WithExecContainer(name string) ExecOption
func WithExecTimeout(d time.Duration) ExecOption

// 错误判断
func IsContainerNotFound(err error) bool
func IsPodNotFound(err error) bool
func IsForbidden(err error) bool
func IsTimeout(err error) bool
func IsConnectionLost(err error) bool
```

### pkg/tenant

```go
// 租户上下文
func WithTenant(ctx context.Context, tenantID string) context.Context
func FromContext(ctx context.Context) string  // 未设置返回 "default"

// 集群上下文
func WithCluster(ctx context.Context, clusterID string) context.Context
func ClusterFromContext(ctx context.Context) string  // 未设置返回 ""

// 组合
func WithTenantAndCluster(ctx context.Context, tenantID, clusterID string) context.Context

// 提取
func ExtractAll(ctx context.Context) (tenantID, clusterID string)
```

### pkg/resource

```go
type Operator struct{}

func NewOperator() *Operator

func (o *Operator) Get(ctx context.Context, cli *client.ClusterClient, obj runtimeclient.Object, key types.NamespacedName) error
func (o *Operator) List(ctx context.Context, cli *client.ClusterClient, list runtimeclient.ObjectList, opts ...ListOption) error
func (o *Operator) Create(ctx context.Context, cli *client.ClusterClient, obj runtimeclient.Object, opts ...CreateOption) error
func (o *Operator) Update(ctx context.Context, cli *client.ClusterClient, obj runtimeclient.Object, opts ...UpdateOption) error
func (o *Operator) Patch(ctx context.Context, cli *client.ClusterClient, obj runtimeclient.Object, patch runtimeclient.Patch, opts ...PatchOption) error
func (o *Operator) Delete(ctx context.Context, cli *client.ClusterClient, obj runtimeclient.Object, opts ...DeleteOption) error

// List 选项
func WithNamespace(ns string) ListOption
func WithLabelSelector(selector labels.Selector) ListOption
func WithFieldSelector(selector fields.Selector) ListOption
func WithLimit(n int64) ListOption
func WithContinue(token string) ListOption
```

### pkg/getter

```go
// 基于 Informer 的缓存读取
type CacheReader struct{}

func NewCacheReader(factory informers.SharedInformerFactory) *CacheReader
func (c *CacheReader) ListPods(namespace string, selector labels.Selector) ([]*corev1.Pod, error)
func (c *CacheReader) GetPod(namespace, name string) (*corev1.Pod, error)
func (c *CacheReader) ListServices(namespace string, selector labels.Selector) ([]*corev1.Service, error)
func (c *CacheReader) ListDeployments(namespace string, selector labels.Selector) ([]*appsv1.Deployment, error)
func (c *CacheReader) HasSynced() bool
func (c *CacheReader) WaitForCacheSync(stopCh <-chan struct{}) bool
```

---

## 设计原则

1. **纯技术** - 无业务逻辑，无租户概念，无配置来源假设
2. **依赖注入友好** - 所有组件可轻松 mock
3. **按需使用** - 只导入需要的包
4. **显式生命周期** - 由使用者控制创建和销毁

---

## 迁移参考

详见 `docs/migration.md`