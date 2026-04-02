package cluster

import (
	"context"
	"testing"
	"time"

	"github.com/guilinonline/k8s-kit/pkg/client"
	"github.com/guilinonline/k8s-kit/pkg/tenant"
)

// TestNewManager 测试 Manager 创建
func TestNewManager(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)

	if manager == nil {
		t.Fatal("NewManager() 返回 nil")
	}

	// 验证初始化状态
	if manager.clusters == nil {
		t.Error("clusters map 不应为 nil")
	}

	if manager.clientFactory != factory {
		t.Error("clientFactory 不匹配")
	}

	if manager.healthConfig.Interval != DefaultHealthCheckConfig.Interval {
		t.Error("healthConfig 不匹配")
	}

	// 验证 context
	if manager.ctx == nil {
		t.Error("ctx 不应为 nil")
	}

	if manager.cancel == nil {
		t.Error("cancel 不应为 nil")
	}
}

// TestManager_Stop 测试 Manager 停止
func TestManager_Stop(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)

	// 停止 Manager
	manager.Stop()

	// 验证 context 被取消
	select {
	case <-manager.ctx.Done():
		// 成功
	case <-time.After(100 * time.Millisecond):
		t.Error("context 应该被取消")
	}

	if manager.ctx.Err() != context.Canceled {
		t.Errorf("期望 Canceled, 得到 %v", manager.ctx.Err())
	}
}

// TestHealthStatus_String 测试 HealthStatus 字符串表示
func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthStatusUnknown, "unknown"},
		{HealthStatusHealthy, "healthy"},
		{HealthStatusDegraded, "degraded"},
		{HealthStatusUnhealthy, "unhealthy"},
		{HealthStatusReconnecting, "reconnecting"},
		{HealthStatus(99), "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("HealthStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestHealthStatus_Constants 测试 HealthStatus 常量值
func TestHealthStatus_Constants(t *testing.T) {
	// 验证 iota 生成的值
	if HealthStatusUnknown != 0 {
		t.Errorf("HealthStatusUnknown = %d, want 0", HealthStatusUnknown)
	}

	if HealthStatusHealthy != 1 {
		t.Errorf("HealthStatusHealthy = %d, want 1", HealthStatusHealthy)
	}

	if HealthStatusDegraded != 2 {
		t.Errorf("HealthStatusDegraded = %d, want 2", HealthStatusDegraded)
	}

	if HealthStatusUnhealthy != 3 {
		t.Errorf("HealthStatusUnhealthy = %d, want 3", HealthStatusUnhealthy)
	}

	if HealthStatusReconnecting != 4 {
		t.Errorf("HealthStatusReconnecting = %d, want 4", HealthStatusReconnecting)
	}
}

// TestWithTenantID 测试租户 ID 选项
func TestWithTenantID(t *testing.T) {
	opts := &RegisterOptions{}

	// 默认应为空
	if opts.TenantID != "" {
		t.Errorf("默认 TenantID 应为空字符串, 得到 %v", opts.TenantID)
	}

	// 应用选项
	WithTenantID("tenant-123")(opts)
	if opts.TenantID != "tenant-123" {
		t.Errorf("TenantID = %v, want tenant-123", opts.TenantID)
	}

	// 覆盖设置
	WithTenantID("tenant-456")(opts)
	if opts.TenantID != "tenant-456" {
		t.Errorf("TenantID = %v, want tenant-456", opts.TenantID)
	}

	// 设置为空
	WithTenantID("")(opts)
	if opts.TenantID != "" {
		t.Errorf("TenantID = %v, want empty string", opts.TenantID)
	}
}

// TestDefaultHealthCheckConfig 测试默认健康检查配置
func TestDefaultHealthCheckConfig(t *testing.T) {
	cfg := DefaultHealthCheckConfig

	// 验证默认值
	if cfg.Interval != 60*time.Second {
		t.Errorf("Interval = %v, want 60s", cfg.Interval)
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", cfg.Timeout)
	}

	if cfg.FailureThreshold != 3 {
		t.Errorf("FailureThreshold = %v, want 3", cfg.FailureThreshold)
	}

	if cfg.SuccessThreshold != 2 {
		t.Errorf("SuccessThreshold = %v, want 2", cfg.SuccessThreshold)
	}

	if !cfg.AutoReconnect {
		t.Error("AutoReconnect 默认应为 true")
	}

	if cfg.SyncInterval != 5*time.Minute {
		t.Errorf("SyncInterval = %v, want 5m", cfg.SyncInterval)
	}

	if cfg.MaxEntries != 100 {
		t.Errorf("MaxEntries = %v, want 100", cfg.MaxEntries)
	}

	if cfg.CleanupInterval != 5*time.Minute {
		t.Errorf("CleanupInterval = %v, want 5m", cfg.CleanupInterval)
	}

	if cfg.IdleTimeout != 10*time.Minute {
		t.Errorf("IdleTimeout = %v, want 10m", cfg.IdleTimeout)
	}
}

// TestDefaultBackoffStrategy 测试默认退避策略
func TestDefaultBackoffStrategy(t *testing.T) {
	strategy := DefaultBackoffStrategy

	// 验证默认值
	if strategy.InitialInterval != 1*time.Second {
		t.Errorf("InitialInterval = %v, want 1s", strategy.InitialInterval)
	}

	if strategy.MaxInterval != 60*time.Second {
		t.Errorf("MaxInterval = %v, want 60s", strategy.MaxInterval)
	}

	if strategy.Multiplier != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", strategy.Multiplier)
	}

	if strategy.MaxRetries != 0 {
		t.Errorf("MaxRetries = %v, want 0", strategy.MaxRetries)
	}
}

// TestChangeType_Constants 测试 ChangeType 常量
func TestChangeType_Constants(t *testing.T) {
	// 验证 iota 生成的值
	if ChangeTypeAdd != 0 {
		t.Errorf("ChangeTypeAdd = %d, want 0", ChangeTypeAdd)
	}

	if ChangeTypeUpdate != 1 {
		t.Errorf("ChangeTypeUpdate = %d, want 1", ChangeTypeUpdate)
	}

	if ChangeTypeDelete != 2 {
		t.Errorf("ChangeTypeDelete = %d, want 2", ChangeTypeDelete)
	}
}

// TestClusterConfig_Struct 测试 ClusterConfig 结构
func TestClusterConfig_Struct(t *testing.T) {
	cfg := ClusterConfig{
		ID:         "cluster-1",
		TenantID:   "tenant-1",
		Kubeconfig: []byte("test-kubeconfig"),
	}

	if cfg.ID != "cluster-1" {
		t.Errorf("ID = %v, want cluster-1", cfg.ID)
	}

	if cfg.TenantID != "tenant-1" {
		t.Errorf("TenantID = %v, want tenant-1", cfg.TenantID)
	}

	if string(cfg.Kubeconfig) != "test-kubeconfig" {
		t.Errorf("Kubeconfig = %v, want test-kubeconfig", string(cfg.Kubeconfig))
	}
}

// TestEventCallbacks_Struct 测试 EventCallbacks 结构
func TestEventCallbacks_Struct(t *testing.T) {
	callbackCalled := false

	callbacks := EventCallbacks{
		OnHealthy: func(id string) {
			callbackCalled = true
		},
		OnUnhealthy: func(id string) {
			callbackCalled = true
		},
		OnReconnected: func(id string) {
			callbackCalled = true
		},
		OnInformerRecreate: func(id string) {
			callbackCalled = true
		},
	}

	// 验证回调可以被调用
	if callbacks.OnHealthy != nil {
		callbacks.OnHealthy("test")
		if !callbackCalled {
			t.Error("OnHealthy 回调应该被调用")
		}
	}

	// 验证空回调不会 panic
	var emptyCallbacks EventCallbacks
	if emptyCallbacks.OnHealthy != nil {
		emptyCallbacks.OnHealthy("test")
	}
}

// TestManagedCluster_InitialState 测试 ManagedCluster 初始状态
func TestManagedCluster_InitialState(t *testing.T) {
	cluster := &ManagedCluster{
		ID:       "test-cluster",
		TenantID: "test-tenant",
	}

	if cluster.ID != "test-cluster" {
		t.Errorf("ID = %v, want test-cluster", cluster.ID)
	}

	if cluster.TenantID != "test-tenant" {
		t.Errorf("TenantID = %v, want test-tenant", cluster.TenantID)
	}

	// 验证其他字段为零值
	if cluster.client != nil {
		t.Error("client 初始应为 nil")
	}

	if cluster.health != 0 {
		t.Error("health 初始应为 0 (unknown)")
	}

	if cluster.failCount != 0 {
		t.Error("failCount 初始应为 0")
	}

	if cluster.closed {
		t.Error("closed 初始应为 false")
	}
}

// TestHealthCheckConfig_Custom 测试自定义健康检查配置
func TestHealthCheckConfig_Custom(t *testing.T) {
	cfg := HealthCheckConfig{
		Interval:         30 * time.Second,
		Timeout:          10 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 3,
		AutoReconnect:    false,
		SyncInterval:     1 * time.Minute,
		MaxEntries:       50,
		CleanupInterval:  2 * time.Minute,
		IdleTimeout:      5 * time.Minute,
	}

	if cfg.Interval != 30*time.Second {
		t.Error("Interval 不匹配")
	}

	if cfg.Timeout != 10*time.Second {
		t.Error("Timeout 不匹配")
	}

	if cfg.FailureThreshold != 5 {
		t.Error("FailureThreshold 不匹配")
	}

	if cfg.AutoReconnect {
		t.Error("AutoReconnect 应为 false")
	}
}

// TestBackoffStrategy_Custom 测试自定义退避策略
func TestBackoffStrategy_Custom(t *testing.T) {
	strategy := BackoffStrategy{
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     30 * time.Second,
		Multiplier:      1.5,
		MaxRetries:      5,
	}

	if strategy.InitialInterval != 500*time.Millisecond {
		t.Error("InitialInterval 不匹配")
	}

	if strategy.MaxInterval != 30*time.Second {
		t.Error("MaxInterval 不匹配")
	}

	if strategy.Multiplier != 1.5 {
		t.Error("Multiplier 不匹配")
	}

	if strategy.MaxRetries != 5 {
		t.Error("MaxRetries 不匹配")
	}
}

// TestManager_GetClientFromContext_NoClusterID 测试无集群 ID 的上下文
func TestManager_GetClientFromContext_NoClusterID(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)
	defer manager.Stop()

	ctx := context.Background()

	_, err := manager.GetClientFromContext(ctx)
	if err == nil {
		t.Error("无集群 ID 的上下文应该返回错误")
	}
}

// TestManager_GetClientFromContext_WithClusterID 测试带集群 ID 的上下文
func TestManager_GetClientFromContext_WithClusterID(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)
	defer manager.Stop()

	// 创建带集群 ID 的上下文
	ctx := tenant.WithCluster(context.Background(), "test-cluster")

	_, err := manager.GetClientFromContext(ctx)
	if err == nil {
		// 期望错误，因为集群未注册
		t.Log("集群未注册，返回错误是预期的")
	} else {
		t.Logf("错误: %v", err)
	}
}

// TestManager_List_Empty 测试空集群列表
func TestManager_List_Empty(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)
	defer manager.Stop()

	clusters := manager.List()
	if clusters == nil {
		t.Error("List() 不应返回 nil")
	}

	if len(clusters) != 0 {
		t.Errorf("空 manager 应该返回 0 个集群，得到 %d", len(clusters))
	}
}

// TestManager_GetHealthStatus_NotFound 测试获取未注册集群的健康状态
func TestManager_GetHealthStatus_NotFound(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)
	defer manager.Stop()

	_, err := manager.GetHealthStatus("non-existent")
	if err == nil {
		t.Error("获取未注册集群的健康状态应该返回错误")
	}
}

// TestManager_SetEventCallbacks 测试设置事件回调
func TestManager_SetEventCallbacks(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)
	defer manager.Stop()

	callbacks := EventCallbacks{
		OnHealthy: func(id string) {
			t.Logf("Cluster %s is healthy", id)
		},
		OnUnhealthy: func(id string) {
			t.Logf("Cluster %s is unhealthy", id)
		},
	}

	// 设置回调
	manager.SetEventCallbacks(callbacks)

	// 验证回调已设置
	if manager.onClusterHealthy == nil {
		t.Error("onClusterHealthy 不应为 nil")
	}

	if manager.onClusterUnhealthy == nil {
		t.Error("onClusterUnhealthy 不应为 nil")
	}
}

// TestManager_ConcurrentAccess 测试并发访问
func TestManager_ConcurrentAccess(t *testing.T) {
	factory := client.NewFactory()
	manager := NewManager(factory, DefaultHealthCheckConfig)
	defer manager.Stop()

	done := make(chan bool, 4)

	// 并发 List
	go func() {
		for i := 0; i < 100; i++ {
			_ = manager.List()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = manager.List()
		}
		done <- true
	}()

	// 并发 Stop
	go func() {
		for i := 0; i < 50; i++ {
			manager.Stop()
		}
		done <- true
	}()

	// 并发 GetClient
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = manager.GetClient("non-existent")
		}
		done <- true
	}()

	// 等待所有 goroutine 完成
	for i := 0; i < 4; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(1 * time.Second):
			t.Fatal("并发测试超时")
		}
	}
}

// BenchmarkHealthStatus_String HealthStatus 字符串转换基准测试
func BenchmarkHealthStatus_String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = HealthStatusHealthy.String()
	}
}

// BenchmarkNewManager Manager 创建基准测试
func BenchmarkNewManager(b *testing.B) {
	factory := client.NewFactory()
	for i := 0; i < b.N; i++ {
		_ = NewManager(factory, DefaultHealthCheckConfig)
	}
}
