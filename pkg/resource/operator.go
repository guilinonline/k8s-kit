package resource

import (
	"context"
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	k8sclient "github.com/guilinonline/k8s-kit/pkg/client"
)

// Operator 资源操作器
// 内部集成 Informer 缓存，Get/List 优先从缓存读取，写操作直接调用 APIServer
type Operator struct {
	mu     sync.RWMutex
	client *k8sclient.ClusterClient

	// Informer 相关（懒加载）
	factory     informers.SharedInformerFactory
	stopCh      chan struct{}
	initialized bool
	initErr     error
}

// NewOperator 创建操作器（无缓存模式）
// 所有操作直接调用 APIServer
func NewOperator() *Operator {
	return &Operator{}
}

// NewOperatorWithClient 创建带缓存的操作器
// Get/List 操作优先从 Informer 缓存读取
// 业务侧无需手动启动 Informer，首次 Get/List 时自动初始化
func NewOperatorWithClient(client *k8sclient.ClusterClient) *Operator {
	return &Operator{
		client: client,
		stopCh: make(chan struct{}),
	}
}

// initCache 懒加载初始化 Informer 缓存
func (o *Operator) initCache() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.initialized {
		return o.initErr
	}

	if o.client == nil || o.client.Clientset == nil {
		o.initErr = fmt.Errorf("client not available")
		o.initialized = true
		return o.initErr
	}

	// 创建 Informer Factory
	o.factory = informers.NewSharedInformerFactory(o.client.Clientset, 0) // 使用默认 resync

	// 启动 Informer
	o.factory.Start(o.stopCh)

	// 等待缓存同步
	o.factory.WaitForCacheSync(o.stopCh)

	o.initialized = true
	return nil
}

// Get 获取资源
// 如果配置了 client 且缓存已初始化，优先从缓存读取 Pod（其他资源类型 fallback 到 API）
func (o *Operator) Get(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	key types.NamespacedName,
) error {
	// 如果是 Pod 且有缓存，使用缓存
	if pod, ok := obj.(*corev1.Pod); ok && o.client != nil {
		if err := o.initCache(); err == nil {
			cachedPod, err := o.factory.Core().V1().Pods().Lister().Pods(key.Namespace).Get(key.Name)
			if err == nil && cachedPod != nil {
				*pod = *cachedPod
				return nil
			}
			// 缓存未命中，fallback 到 API
		}
	}

	// 使用 RuntimeClient 直接调用 API
	return cli.RuntimeClient.Get(ctx, key, obj)
}

// List 列表查询
// 如果配置了 client 且缓存已初始化，优先从缓存读取 Pod/Service/Deployment
func (o *Operator) List(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	list runtimeclient.ObjectList,
	opts ...ListOption,
) error {
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 如果是 PodList 且有缓存，使用缓存
	if podList, ok := list.(*corev1.PodList); ok && o.client != nil {
		if err := o.initCache(); err == nil {
			var selector labels.Selector
			if options.LabelSelector != nil {
				selector = options.LabelSelector
			} else {
				selector = labels.Everything()
			}

			cachedPods, err := o.factory.Core().V1().Pods().Lister().Pods(options.Namespace).List(selector)
			if err == nil {
				podList.Items = make([]corev1.Pod, len(cachedPods))
				for i, p := range cachedPods {
					podList.Items[i] = *p
				}
				return nil
			}
		}
	}

	// 使用 RuntimeClient 直接调用 API
	listOpts := []runtimeclient.ListOption{}
	if options.Namespace != "" {
		listOpts = append(listOpts, runtimeclient.InNamespace(options.Namespace))
	}
	if options.LabelSelector != nil {
		listOpts = append(listOpts, runtimeclient.MatchingLabelsSelector{Selector: options.LabelSelector})
	}
	if options.FieldSelector != nil {
		listOpts = append(listOpts, runtimeclient.MatchingFieldsSelector{Selector: options.FieldSelector})
	}
	if options.Limit > 0 {
		listOpts = append(listOpts, runtimeclient.Limit(options.Limit))
	}
	if options.Continue != "" {
		listOpts = append(listOpts, runtimeclient.Continue(options.Continue))
	}

	return cli.RuntimeClient.List(ctx, list, listOpts...)
}

// Create 创建资源 - 直接调用 APIServer
func (o *Operator) Create(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	opts ...CreateOption,
) error {
	options := &CreateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	createOpts := []runtimeclient.CreateOption{}
	if options.FieldManager != "" {
		createOpts = append(createOpts, runtimeclient.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Create(ctx, obj, createOpts...)
}

// Update 更新资源 - 直接调用 APIServer
func (o *Operator) Update(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	opts ...UpdateOption,
) error {
	options := &UpdateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	updateOpts := []runtimeclient.UpdateOption{}
	if options.FieldManager != "" {
		updateOpts = append(updateOpts, runtimeclient.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Update(ctx, obj, updateOpts...)
}

// Patch 补丁更新 - 直接调用 APIServer
func (o *Operator) Patch(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	patch runtimeclient.Patch,
	opts ...PatchOption,
) error {
	options := &PatchOptions{}
	for _, opt := range opts {
		opt(options)
	}

	patchOpts := []runtimeclient.PatchOption{}
	if options.FieldManager != "" {
		patchOpts = append(patchOpts, runtimeclient.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Patch(ctx, obj, patch, patchOpts...)
}

// Delete 删除资源 - 直接调用 APIServer
func (o *Operator) Delete(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	opts ...DeleteOption,
) error {
	options := &DeleteOptions{}
	for _, opt := range opts {
		opt(options)
	}

	deleteOpts := []runtimeclient.DeleteOption{}
	if options.GracePeriodSeconds != nil {
		deleteOpts = append(deleteOpts, runtimeclient.GracePeriodSeconds(*options.GracePeriodSeconds))
	}

	return cli.RuntimeClient.Delete(ctx, obj, deleteOpts...)
}

// Stop 停止 Operator，清理 Informer 资源
// 在程序退出或不再需要缓存时调用
func (o *Operator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.stopCh != nil {
		close(o.stopCh)
		o.stopCh = nil
	}
	o.initialized = false
}

// IsCached 检查缓存是否已初始化
func (o *Operator) IsCached() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.initialized
}
