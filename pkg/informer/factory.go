package informer

import (
	//"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/informers"
	//"k8s.io/client-go/kubernetes"

	"github.com/guilinonline/k8s-kit/pkg/client"
)

// Factory Informer工厂
type Factory struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	config  FactoryConfig
}

// FactoryConfig 工厂配置
type FactoryConfig struct {
	MaxEntries      int           // 最大Informer数量
	CleanupInterval time.Duration // 清理间隔
	IdleTimeout     time.Duration // 空闲超时
}

// DefaultFactoryConfig 默认配置
var DefaultFactoryConfig = FactoryConfig{
	MaxEntries:      100,
	CleanupInterval: 5 * time.Minute,
	IdleTimeout:     10 * time.Minute,
}

// NewFactory 创建Informer工厂
func NewFactory(config ...FactoryConfig) *Factory {
	cfg := DefaultFactoryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	f := &Factory{
		entries: make(map[string]*Entry),
		config:  cfg,
	}

	// 启动清理协程
	go f.cleanupLoop()

	return f
}

// Create 创建Informer
func (f *Factory) Create(
	client *client.ClusterClient,
	opts Options,
) (*Entry, error) {
	key := generateKey(client, opts.Namespace)

	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查是否已存在
	if entry, ok := f.entries[key]; ok {
		entry.UpdateAccessTime()
		return entry, nil
	}

	// 创建新的Informer
	entry, err := f.createEntry(client, opts, key)
	if err != nil {
		return nil, err
	}

	f.entries[key] = entry
	return entry, nil
}

// Get 获取已存在的Informer
func (f *Factory) Get(key string) (*Entry, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	entry, ok := f.entries[key]
	if ok {
		entry.UpdateAccessTime()
	}
	return entry, ok
}

// GetOrCreate 获取或创建
func (f *Factory) GetOrCreate(
	key string,
	client *client.ClusterClient,
	opts Options,
) (*Entry, error) {
	if entry, ok := f.Get(key); ok {
		return entry, nil
	}
	return f.Create(client, opts)
}

// Stop 停止指定Informer
func (f *Factory) Stop(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry, ok := f.entries[key]
	if !ok {
		return fmt.Errorf("informer not found: %s", key)
	}

	entry.Stop()
	delete(f.entries, key)
	return nil
}

// StopAll 停止所有Informer
func (f *Factory) StopAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, entry := range f.entries {
		entry.Stop()
	}
	f.entries = make(map[string]*Entry)
}

// createEntry 创建Informer条目
func (f *Factory) createEntry(
	client *client.ClusterClient,
	opts Options,
	key string,
) (*Entry, error) {
	// 设置默认值
	if opts.ResyncPeriod == 0 {
		opts.ResyncPeriod = DefaultOptions.ResyncPeriod
	}

	// 创建Informer选项
	var informerOpts []informers.SharedInformerOption
	if opts.Namespace != "" {
		informerOpts = append(informerOpts, informers.WithNamespace(opts.Namespace))
	}

	// 创建SharedInformerFactory
	factory := informers.NewSharedInformerFactoryWithOptions(
		client.Clientset,
		opts.ResyncPeriod,
		informerOpts...,
	)

	// 启动Informer
	stopCh := make(chan struct{})
	factory.Start(stopCh)

	// 等待同步
	factory.WaitForCacheSync(make(chan struct{}))

	return &Entry{
		Key:          key,
		Factory:      factory,
		StopCh:       stopCh,
		Lifecycle:    opts.Lifecycle,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}, nil
}

// cleanupLoop 清理循环
func (f *Factory) cleanupLoop() {
	ticker := time.NewTicker(f.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		f.cleanup()
	}
}

// cleanup 清理过期Informer
func (f *Factory) cleanup() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	for key, entry := range f.entries {
		// 只清理按需创建的Informer
		if entry.Lifecycle == LifecycleOnDemand {
			if now.Sub(entry.LastAccessed) > f.config.IdleTimeout {
				entry.Stop()
				delete(f.entries, key)
			}
		}
	}

	// 如果超过最大数量，清理最久未使用的
	if len(f.entries) > f.config.MaxEntries {
		// TODO: 实现LRU淘汰
	}
}

// generateKey 生成Informer key
func generateKey(client *client.ClusterClient, namespace string) string {
	// 使用RESTConfig的Host作为集群标识
	host := ""
	if client.RESTConfig != nil {
		host = client.RESTConfig.Host
	}
	if namespace == "" {
		return host
	}
	return fmt.Sprintf("%s/%s", host, namespace)
}
