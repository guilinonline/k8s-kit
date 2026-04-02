package client

import (
	"testing"
	"time"

	"k8s.io/client-go/rest"
)

// TestNewFactory 测试工厂创建
func TestNewFactory(t *testing.T) {
	f := NewFactory()

	if f == nil {
		t.Fatal("NewFactory() 返回 nil")
	}

	// 验证默认值
	if f.defaultOptions.Timeout != DefaultTimeout {
		t.Errorf("默认 Timeout = %v, want %v", f.defaultOptions.Timeout, DefaultTimeout)
	}

	if f.defaultOptions.QPS != DefaultQPS {
		t.Errorf("默认 QPS = %v, want %v", f.defaultOptions.QPS, DefaultQPS)
	}

	if f.defaultOptions.Burst != DefaultBurst {
		t.Errorf("默认 Burst = %v, want %v", f.defaultOptions.Burst, DefaultBurst)
	}
}

// TestNewFactory_WithOptions 测试带选项的工厂创建
func TestNewFactory_WithOptions(t *testing.T) {
	f := NewFactory(
		WithTimeout(60*time.Second),
		WithQPS(100),
		WithBurst(200),
		WithUserAgent("test-agent"),
	)

	if f == nil {
		t.Fatal("NewFactory() 返回 nil")
	}

	// 验证选项值
	if f.defaultOptions.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want %v", f.defaultOptions.Timeout, 60*time.Second)
	}

	if f.defaultOptions.QPS != 100 {
		t.Errorf("QPS = %v, want 100", f.defaultOptions.QPS)
	}

	if f.defaultOptions.Burst != 200 {
		t.Errorf("Burst = %v, want 200", f.defaultOptions.Burst)
	}

	if f.defaultOptions.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %v, want test-agent", f.defaultOptions.UserAgent)
	}
}

// TestWithTimeout 测试超时选项
func TestWithTimeout(t *testing.T) {
	opts := &ClientOptions{}

	WithTimeout(30 * time.Second)(opts)
	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", opts.Timeout)
	}

	WithTimeout(0)(opts)
	if opts.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0", opts.Timeout)
	}

	WithTimeout(-1 * time.Second)(opts)
	if opts.Timeout != -1*time.Second {
		t.Errorf("Timeout = %v, want -1s", opts.Timeout)
	}
}

// TestWithQPS 测试 QPS 选项
func TestWithQPS(t *testing.T) {
	opts := &ClientOptions{}

	WithQPS(50)(opts)
	if opts.QPS != 50 {
		t.Errorf("QPS = %v, want 50", opts.QPS)
	}

	WithQPS(0)(opts)
	if opts.QPS != 0 {
		t.Errorf("QPS = %v, want 0", opts.QPS)
	}

	WithQPS(-10)(opts)
	if opts.QPS != -10 {
		t.Errorf("QPS = %v, want -10", opts.QPS)
	}

	WithQPS(9999.99)(opts)
	if opts.QPS != 9999.99 {
		t.Errorf("QPS = %v, want 9999.99", opts.QPS)
	}
}

// TestWithBurst 测试 Burst 选项
func TestWithBurst(t *testing.T) {
	opts := &ClientOptions{}

	WithBurst(100)(opts)
	if opts.Burst != 100 {
		t.Errorf("Burst = %v, want 100", opts.Burst)
	}

	WithBurst(0)(opts)
	if opts.Burst != 0 {
		t.Errorf("Burst = %v, want 0", opts.Burst)
	}

	WithBurst(-5)(opts)
	if opts.Burst != -5 {
		t.Errorf("Burst = %v, want -5", opts.Burst)
	}
}

// TestWithUserAgent 测试 UserAgent 选项
func TestWithUserAgent(t *testing.T) {
	opts := &ClientOptions{}

	WithUserAgent("my-agent/1.0")(opts)
	if opts.UserAgent != "my-agent/1.0" {
		t.Errorf("UserAgent = %v, want my-agent/1.0", opts.UserAgent)
	}

	WithUserAgent("")(opts)
	if opts.UserAgent != "" {
		t.Errorf("UserAgent = %v, want empty string", opts.UserAgent)
	}
}

// TestWithImpersonate 测试 Impersonate 选项
func TestWithImpersonate(t *testing.T) {
	opts := &ClientOptions{}

	// 默认应为 nil
	if opts.Impersonate != nil {
		t.Error("默认 Impersonate 应为 nil")
	}

	// 设置 Impersonate
	impersonateConfig := &rest.ImpersonationConfig{
		UserName: "admin@example.com",
		Groups:   []string{"system:masters"},
	}
	opts.Impersonate = impersonateConfig

	if opts.Impersonate != impersonateConfig {
		t.Error("Impersonate 不匹配")
	}

	if opts.Impersonate.UserName != "admin@example.com" {
		t.Errorf("UserName = %v, want admin@example.com", opts.Impersonate.UserName)
	}
}

// TestClientOptions_Default 测试 ClientOptions 默认值
func TestClientOptions_Default(t *testing.T) {
	opts := &ClientOptions{}

	// 所有字段应为零值
	if opts.Timeout != 0 {
		t.Errorf("Timeout 默认应为 0, 得到 %v", opts.Timeout)
	}

	if opts.QPS != 0 {
		t.Errorf("QPS 默认应为 0, 得到 %v", opts.QPS)
	}

	if opts.Burst != 0 {
		t.Errorf("Burst 默认应为 0, 得到 %v", opts.Burst)
	}

	if opts.UserAgent != "" {
		t.Errorf("UserAgent 默认应为空字符串, 得到 %v", opts.UserAgent)
	}

	if opts.Impersonate != nil {
		t.Error("Impersonate 默认应为 nil")
	}
}

// TestClusterClient_Struct 测试 ClusterClient 结构
func TestClusterClient_Struct(t *testing.T) {
	// 创建空的 ClusterClient
	c := &ClusterClient{}

	// 所有字段应为 nil
	if c.RESTConfig != nil {
		t.Error("RESTConfig 应为 nil")
	}

	if c.Clientset != nil {
		t.Error("Clientset 应为 nil")
	}

	if c.RuntimeClient != nil {
		t.Error("RuntimeClient 应为 nil")
	}
}

// TestDefaultConstants 测试默认常量
func TestDefaultConstants(t *testing.T) {
	// 验证默认值
	if DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v, want 30s", DefaultTimeout)
	}

	if DefaultQPS != 50 {
		t.Errorf("DefaultQPS = %v, want 50", DefaultQPS)
	}

	if DefaultBurst != 100 {
		t.Errorf("DefaultBurst = %v, want 100", DefaultBurst)
	}
}

// TestCreateFromKubeconfig_InvalidKubeconfig 测试无效的 kubeconfig
func TestCreateFromKubeconfig_InvalidKubeconfig(t *testing.T) {
	f := NewFactory()

	// 测试无效的 kubeconfig
	invalidKubeconfig := []byte("invalid yaml content")

	_, err := f.CreateFromKubeconfig(invalidKubeconfig)
	if err == nil {
		t.Error("无效的 kubeconfig 应该返回错误")
	}
}

// TestCreateFromKubeconfig_EmptyKubeconfig 测试空的 kubeconfig
func TestCreateFromKubeconfig_EmptyKubeconfig(t *testing.T) {
	f := NewFactory()

	// 测试空的 kubeconfig
	emptyKubeconfig := []byte("")

	_, err := f.CreateFromKubeconfig(emptyKubeconfig)
	if err == nil {
		t.Error("空的 kubeconfig 应该返回错误")
	}
}

// TestFactory_MultipleCalls 测试工厂多次调用
func TestFactory_MultipleCalls(t *testing.T) {
	f := NewFactory(
		WithTimeout(30*time.Second),
		WithQPS(50),
	)

	// 验证工厂配置在多调用间保持一致
	if f.defaultOptions.Timeout != 30*time.Second {
		t.Error("Timeout 不匹配")
	}

	if f.defaultOptions.QPS != 50 {
		t.Error("QPS 不匹配")
	}
}

// TestOption_FunctionSignature 测试 Option 函数签名
func TestOption_FunctionSignature(t *testing.T) {
	// 验证 Option 类型可以正确应用
	var opt Option = WithTimeout(10 * time.Second)

	opts := &ClientOptions{}
	opt(opts)

	if opts.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", opts.Timeout)
	}
}

// TestClientOptions_MultipleOptions 测试多个选项应用
func TestClientOptions_MultipleOptions(t *testing.T) {
	opts := &ClientOptions{}

	// 应用多个选项
	options := []Option{
		WithTimeout(60 * time.Second),
		WithQPS(100),
		WithBurst(200),
		WithUserAgent("test-agent"),
	}

	for _, opt := range options {
		opt(opts)
	}

	// 验证所有选项都已应用
	if opts.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", opts.Timeout)
	}

	if opts.QPS != 100 {
		t.Errorf("QPS = %v, want 100", opts.QPS)
	}

	if opts.Burst != 200 {
		t.Errorf("Burst = %v, want 200", opts.Burst)
	}

	if opts.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %v, want test-agent", opts.UserAgent)
	}
}

// TestClientOptions_Override 测试选项覆盖
func TestClientOptions_Override(t *testing.T) {
	opts := &ClientOptions{}

	// 先设置一个值
	WithTimeout(30 * time.Second)(opts)
	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", opts.Timeout)
	}

	// 覆盖设置
	WithTimeout(60 * time.Second)(opts)
	if opts.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", opts.Timeout)
	}

	// 再次覆盖
	WithTimeout(0)(opts)
	if opts.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0", opts.Timeout)
	}
}

// BenchmarkNewFactory 工厂创建基准测试
func BenchmarkNewFactory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewFactory(
			WithTimeout(30*time.Second),
			WithQPS(50),
			WithBurst(100),
		)
	}
}

// BenchmarkWithTimeout Timeout 选项基准测试
func BenchmarkWithTimeout(b *testing.B) {
	opts := &ClientOptions{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithTimeout(30 * time.Second)(opts)
	}
}
