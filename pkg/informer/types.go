package informer

import (
	"sync"
	"time"

	"k8s.io/client-go/informers"
)

// Options Informer选项
type Options struct {
	Namespace    string
	ResyncPeriod time.Duration
	Lifecycle    Lifecycle
}

// Lifecycle Informer生命周期
type Lifecycle int

const (
	// LifecyclePersistent 长期复用
	LifecyclePersistent Lifecycle = iota
	// LifecycleOnDemand 按需创建
	LifecycleOnDemand
	// LifecycleManual 手动管理
	LifecycleManual
)

// DefaultOptions 默认选项
var DefaultOptions = Options{
	Namespace:    "",
	ResyncPeriod: 5 * time.Minute, // 调大 Resync 间隔，减少 APIServer 压力
	Lifecycle:    LifecyclePersistent,
}

// Entry Informer条目
type Entry struct {
	Key          string
	Factory      informers.SharedInformerFactory
	StopCh       chan struct{}
	Lifecycle    Lifecycle
	CreatedAt    time.Time
	LastAccessed time.Time
	RefCount     int32
	mu           sync.RWMutex
}

// Stop 停止Informer
func (e *Entry) Stop() {
	close(e.StopCh)
}

// UpdateAccessTime 更新访问时间
func (e *Entry) UpdateAccessTime() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.LastAccessed = time.Now()
}
