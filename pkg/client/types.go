package client

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterClient K8s客户端集合
type ClusterClient struct {
	RESTConfig    *rest.Config
	Clientset     *kubernetes.Clientset
	RuntimeClient client.Client
}

// ClientOptions 客户端选项
type ClientOptions struct {
	Timeout     time.Duration
	QPS         float32
	Burst       int
	UserAgent   string
	Impersonate *rest.ImpersonationConfig
}

// Option 客户端选项函数
type Option func(*ClientOptions)

// WithTimeout 设置超时
func WithTimeout(timeout time.Duration) Option {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// WithQPS 设置QPS
func WithQPS(qps float32) Option {
	return func(o *ClientOptions) {
		o.QPS = qps
	}
}

// WithBurst 设置Burst
func WithBurst(burst int) Option {
	return func(o *ClientOptions) {
		o.Burst = burst
	}
}

// WithUserAgent 设置UserAgent
func WithUserAgent(ua string) Option {
	return func(o *ClientOptions) {
		o.UserAgent = ua
	}
}
