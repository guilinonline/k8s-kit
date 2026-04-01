package getter

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/seaman/k8s-kit/pkg/informer"
)

// CacheReader 基于Informer的缓存读取器
type CacheReader struct {
	factory informers.SharedInformerFactory
	stopCh  chan struct{}
}

// NewCacheReader 创建缓存读取器
func NewCacheReader(factory informers.SharedInformerFactory) *CacheReader {
	return &CacheReader{
		factory: factory,
		stopCh:  make(chan struct{}),
	}
}

// ListPods 从缓存获取Pod列表
func (c *CacheReader) ListPods(namespace string, selector labels.Selector) ([]interface{}, error) {
	podLister := c.factory.Core().V1().Pods().Lister()
	
	if namespace != "" {
		return podLister.Pods(namespace).List(selector)
	}
	return podLister.List(selector)
}

// GetPod 从缓存获取单个Pod
func (c *CacheReader) GetPod(namespace, name string) (interface{}, error) {
	return c.factory.Core().V1().Pods().Lister().Pods(namespace).Get(name)
}

// ListServices 从缓存获取Service列表
func (c *CacheReader) ListServices(namespace string, selector labels.Selector) ([]interface{}, error) {
	svcLister := c.factory.Core().V1().Services().Lister()
	
	if namespace != "" {
		return svcLister.Services(namespace).List(selector)
	}
	return svcLister.List(selector)
}

// ListDeployments 从缓存获取Deployment列表
func (c *CacheReader) ListDeployments(namespace string, selector labels.Selector) ([]interface{}, error) {
	depLister := c.factory.Apps().V1().Deployments().Lister()
	
	if namespace != "" {
		return depLister.Deployments(namespace).List(selector)
	}
	return depLister.List(selector)
}

// HasSynced 检查缓存是否已同步
func (c *CacheReader) HasSynced() bool {
	return c.factory.Core().V1().Pods().Informer().HasSynced() &&
		c.factory.Core().V1().Services().Informer().HasSynced()
}

// WaitForCacheSync 等待缓存同步
func (c *CacheReader) WaitForCacheSync(stopCh <-chan struct{}) bool {
	return cache.WaitForCacheSync(stopCh,
		c.factory.Core().V1().Pods().Informer().HasSynced,
		c.factory.Core().V1().Services().Informer().HasSynced,
	)
}

// CacheGetter 基于Informer的Getter
type CacheGetter struct {
	reader *CacheReader
}

// NewCacheGetter 创建缓存Getter
func NewCacheGetter(reader *CacheReader) *CacheGetter {
	return &CacheGetter{reader: reader}
}

// GetPod 从缓存获取Pod
func (g *CacheGetter) GetPod(namespace, name string) (*corev1.Pod, error) {
	obj, err := g.reader.GetPod(namespace, name)
	if err != nil {
		return nil, err
	}
	return obj.(*corev1.Pod), nil
}
