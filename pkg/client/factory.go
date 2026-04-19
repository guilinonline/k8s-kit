package client

import (
	"time"

	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"sigs.k8s.io/controller-runtime/pkg/cluster"

	"github.com/pkg/errors"
)

// Factory 客户端工厂
type Factory struct {
	defaultOptions ClientOptions
}

// NewFactory 创建工厂
func NewFactory(opts ...Option) *Factory {
	f := &Factory{
		defaultOptions: ClientOptions{
			Timeout: DefaultTimeout,
			QPS:     DefaultQPS,
			Burst:   DefaultBurst,
		},
	}
	for _, opt := range opts {
		opt(&f.defaultOptions)
	}
	return f
}

const (
	DefaultTimeout = 30 * time.Second
	DefaultQPS     = 50  // 降低默认 QPS，避免打挂 APIServer
	DefaultBurst   = 100 // 降低默认 Burst
)

// CreateFromKubeconfig 从kubeconfig创建
func (f *Factory) CreateFromKubeconfig(kubeconfig []byte, opts ...Option) (*ClusterClient, error) {
	restConf, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create rest config from kubeconfig")
	}
	return f.createFromRESTConfig(restConf, opts...)
}

// CreateFromRESTConfig 从REST配置创建
func (f *Factory) CreateFromRESTConfig(restConf *rest.Config, opts ...Option) (*ClusterClient, error) {
	return f.createFromRESTConfig(restConf, opts...)
}

// createFromRESTConfig 内部实现
func (f *Factory) createFromRESTConfig(restConf *rest.Config, opts ...Option) (*ClusterClient, error) {
	// 合并选项
	options := f.defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	// 设置选项
	restConf.Timeout = options.Timeout
	restConf.QPS = options.QPS
	restConf.Burst = options.Burst
	if options.UserAgent != "" {
		restConf.UserAgent = options.UserAgent
	}
	if options.Impersonate != nil {
		restConf.Impersonate = *options.Impersonate
	}
	if options.DialContext != nil {
		restConf.Dial = options.DialContext
	}

	c := &ClusterClient{
		RESTConfig: restConf,
	}

	// 创建Clientset
	clientset, err := kubernetes.NewForConfig(restConf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset")
	}
	c.Clientset = clientset

	// 创建controller-runtime client
	rtClient, err := client.New(restConf, client.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create runtime client")
	}
	c.RuntimeClient = rtClient

	return c, nil
}

// CreateFromInCluster 从集群内创建
func (f *Factory) CreateFromInCluster(opts ...Option) (*ClusterClient, error) {
	restConf, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get in-cluster config")
	}
	return f.createFromRESTConfig(restConf, opts...)
}
