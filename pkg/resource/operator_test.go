package resource

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestNewOperator 测试无缓存模式的 Operator 创建
func TestNewOperator(t *testing.T) {
	op := NewOperator()

	if op == nil {
		t.Fatal("NewOperator() 返回 nil")
	}

	// 检查默认值
	if op.client != nil {
		t.Error("无缓存模式的 client 应为 nil")
	}

	if op.initialized {
		t.Error("新创建的 Operator initialized 应为 false")
	}

	if op.stopCh != nil {
		t.Error("无缓存模式的 stopCh 应为 nil")
	}

	// IsCached 应该返回 false
	if op.IsCached() {
		t.Error("无缓存模式 IsCached() 应返回 false")
	}
}

// TestNewOperatorWithClient 测试带缓存模式的 Operator 创建
func TestNewOperatorWithClient(t *testing.T) {
	// 注意：这里不传入真实的 client，只测试结构
	op := &Operator{
		stopCh: make(chan struct{}),
	}

	if op == nil {
		t.Fatal("Operator 创建失败")
	}

	if op.stopCh == nil {
		t.Error("带缓存模式的 stopCh 不应为 nil")
	}

	// IsCached 在还没有初始化缓存时应该返回 false
	if op.IsCached() {
		t.Error("未初始化的 Operator IsCached() 应返回 false")
	}
}

// TestOperator_IsCached 测试缓存状态检查
func TestOperator_IsCached(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
		expected    bool
	}{
		{
			name:        "未初始化",
			initialized: false,
			expected:    false,
		},
		{
			name:        "已初始化",
			initialized: true,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &Operator{
				initialized: tt.initialized,
			}

			if got := op.IsCached(); got != tt.expected {
				t.Errorf("IsCached() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestOperator_Stop 测试停止 Operator
func TestOperator_Stop(t *testing.T) {
	t.Run("停止带缓存的 Operator", func(t *testing.T) {
		stopCh := make(chan struct{})
		op := &Operator{
			stopCh:      stopCh,
			initialized: true,
		}

		// 在 goroutine 中停止，避免阻塞
		go func() {
			time.Sleep(10 * time.Millisecond)
			op.Stop()
		}()

		// 验证 channel 被关闭
		done := make(chan bool)
		go func() {
			<-stopCh
			done <- true
		}()

		select {
		case <-done:
			// 成功
		case <-time.After(100 * time.Millisecond):
			t.Error("stopCh 没有被关闭")
		}

		// 验证状态重置
		if op.initialized {
			t.Error("Stop() 后 initialized 应为 false")
		}

		if op.stopCh != nil {
			t.Error("Stop() 后 stopCh 应为 nil")
		}
	})

	t.Run("停止无缓存的 Operator", func(t *testing.T) {
		op := NewOperator()

		// 不应 panic
		op.Stop()

		if op.stopCh != nil {
			t.Error("无缓存 Operator Stop() 后 stopCh 应为 nil")
		}
	})

	t.Run("重复停止", func(t *testing.T) {
		stopCh := make(chan struct{})
		op := &Operator{
			stopCh:      stopCh,
			initialized: true,
		}

		// 第一次停止
		op.Stop()

		// 第二次停止不应 panic
		op.Stop()
	})
}

// TestOperator_Stop_Concurrent 测试并发停止
func TestOperator_Stop_Concurrent(t *testing.T) {
	stopCh := make(chan struct{})
	op := &Operator{
		stopCh:      stopCh,
		initialized: true,
	}

	// 并发调用 Stop
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			defer func() {
				// 捕获可能的 panic
				recover()
				done <- true
			}()
			op.Stop()
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(100 * time.Millisecond):
			t.Error("并发 Stop 超时")
		}
	}
}

// TestOperator_IsCached_Concurrent 测试并发缓存状态检查
func TestOperator_IsCached_Concurrent(t *testing.T) {
	op := &Operator{
		initialized: false,
	}

	done := make(chan bool, 2)

	// 并发读取
	go func() {
		for i := 0; i < 100; i++ {
			_ = op.IsCached()
		}
		done <- true
	}()

	// 并发写入
	go func() {
		for i := 0; i < 100; i++ {
			op.mu.Lock()
			op.initialized = !op.initialized
			op.mu.Unlock()
		}
		done <- true
	}()

	// 等待完成
	for i := 0; i < 2; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(1 * time.Second):
			t.Error("并发测试超时")
		}
	}
}

// TestListOptions_Apply 测试 ListOptions 应用
func TestListOptions_Apply(t *testing.T) {
	opts := &ListOptions{}

	// 应用所有选项
	listOpts := []ListOption{
		WithNamespace("default"),
		WithLimit(100),
		WithContinue("token123"),
	}

	for _, opt := range listOpts {
		opt(opts)
	}

	// 验证
	if opts.Namespace != "default" {
		t.Errorf("Namespace = %s, want default", opts.Namespace)
	}

	if opts.Limit != 100 {
		t.Errorf("Limit = %d, want 100", opts.Limit)
	}

	if opts.Continue != "token123" {
		t.Errorf("Continue = %s, want token123", opts.Continue)
	}
}

// TestCreateOptions_Apply 测试 CreateOptions 应用
func TestCreateOptions_Apply(t *testing.T) {
	opts := &CreateOptions{}

	WithFieldManager("my-controller")(opts)

	if opts.FieldManager != "my-controller" {
		t.Errorf("FieldManager = %s, want my-controller", opts.FieldManager)
	}
}

// TestUpdateOptions_Apply 测试 UpdateOptions 应用
func TestUpdateOptions_Apply(t *testing.T) {
	opts := &UpdateOptions{}

	WithUpdateFieldManager("updater")(opts)

	if opts.FieldManager != "updater" {
		t.Errorf("FieldManager = %s, want updater", opts.FieldManager)
	}
}

// TestPatchOptions_Apply 测试 PatchOptions 应用
func TestPatchOptions_Apply(t *testing.T) {
	opts := &PatchOptions{}

	patchOpts := []PatchOption{
		WithPatchFieldManager("patcher"),
		WithForce(true),
	}

	for _, opt := range patchOpts {
		opt(opts)
	}

	if opts.FieldManager != "patcher" {
		t.Errorf("FieldManager = %s, want patcher", opts.FieldManager)
	}

	if opts.Force == nil || !*opts.Force {
		t.Error("Force should be true")
	}
}

// TestDeleteOptions_Apply 测试 DeleteOptions 应用
func TestDeleteOptions_Apply(t *testing.T) {
	opts := &DeleteOptions{}

	WithGracePeriodSeconds(60)(opts)

	if opts.GracePeriodSeconds == nil || *opts.GracePeriodSeconds != 60 {
		t.Errorf("GracePeriodSeconds = %v, want 60", opts.GracePeriodSeconds)
	}
}

// TestNamespacedName 测试 types.NamespacedName 的使用
func TestNamespacedName(t *testing.T) {
	tests := []struct {
		name     string
		key      types.NamespacedName
		wantName string
		wantNS   string
	}{
		{
			name:     "有命名空间",
			key:      types.NamespacedName{Name: "my-pod", Namespace: "default"},
			wantName: "my-pod",
			wantNS:   "default",
		},
		{
			name:     "无命名空间（集群范围资源）",
			key:      types.NamespacedName{Name: "my-node"},
			wantName: "my-node",
			wantNS:   "",
		},
		{
			name:     "空值",
			key:      types.NamespacedName{},
			wantName: "",
			wantNS:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", tt.key.Name, tt.wantName)
			}
			if tt.key.Namespace != tt.wantNS {
				t.Errorf("Namespace = %s, want %s", tt.key.Namespace, tt.wantNS)
			}
		})
	}
}

// TestPodDeepCopy 测试 Pod 深拷贝（用于理解缓存行为）
func TestPodDeepCopy(t *testing.T) {
	original := &corev1.Pod{
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
					Name:  "main",
					Image: "nginx:latest",
				},
			},
		},
	}

	// 深拷贝
	copied := original.DeepCopy()

	// 验证是不同对象
	if copied == original {
		t.Error("DeepCopy 应该返回新对象")
	}

	// 验证值相等
	if copied.Name != original.Name {
		t.Error("深拷贝后的 Name 应该相等")
	}

	// 修改副本不应影响原对象
	copied.Labels["app"] = "modified"
	if original.Labels["app"] == "modified" {
		t.Error("修改副本不应影响原对象")
	}
}

// TestContextPropagation 测试上下文传递
func TestContextPropagation(t *testing.T) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 验证上下文
	if ctx == nil {
		t.Error("context 不应为 nil")
	}

	// 验证超时
	time.Sleep(150 * time.Millisecond)
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("期望 DeadlineExceeded，得到 %v", ctx.Err())
	}
}

// TestContextCancellation 测试上下文取消
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	go func() {
		<-ctx.Done()
		done <- true
	}()

	// 取消上下文
	cancel()

	select {
	case <-done:
		// 成功
	case <-time.After(100 * time.Millisecond):
		t.Error("上下文取消未传播")
	}

	if ctx.Err() != context.Canceled {
		t.Errorf("期望 Canceled，得到 %v", ctx.Err())
	}
}

// TestEmptyPodList 测试空 PodList
func TestEmptyPodList(t *testing.T) {
	podList := &corev1.PodList{}

	if podList.Items == nil {
		t.Log("Items 初始为 nil")
	}

	if len(podList.Items) != 0 {
		t.Error("空 PodList 应该没有 Items")
	}

	// 添加一个 Pod
	podList.Items = append(podList.Items, corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	})

	if len(podList.Items) != 1 {
		t.Error("应该有一个 Item")
	}
}

// TestMatchTypeValues 测试 MatchType 常量值
func TestMatchTypeValues(t *testing.T) {
	// 验证 iota 生成的值
	if MatchContains != 0 {
		t.Errorf("MatchContains = %d, want 0", MatchContains)
	}

	if MatchPrefix != 1 {
		t.Errorf("MatchPrefix = %d, want 1", MatchPrefix)
	}

	if MatchSuffix != 2 {
		t.Errorf("MatchSuffix = %d, want 2", MatchSuffix)
	}

	if MatchRegex != 3 {
		t.Errorf("MatchRegex = %d, want 3", MatchRegex)
	}
}

// TestOperator_ThreadSafety 测试 Operator 的线程安全性
func TestOperator_ThreadSafety(t *testing.T) {
	op := NewOperator()

	done := make(chan bool, 4)

	// 并发读取 IsCached
	go func() {
		for i := 0; i < 100; i++ {
			_ = op.IsCached()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = op.IsCached()
		}
		done <- true
	}()

	// 并发 Stop
	go func() {
		for i := 0; i < 50; i++ {
			op.Stop()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			op.Stop()
		}
		done <- true
	}()

	// 等待所有 goroutine 完成
	for i := 0; i < 4; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(1 * time.Second):
			t.Fatal("线程安全测试超时")
		}
	}
}

// TestGracePeriodEdgeCases 测试优雅期边界情况
func TestGracePeriodEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int64
		expected int64
	}{
		{
			name:     "立即删除",
			seconds:  0,
			expected: 0,
		},
		{
			name:     "30秒",
			seconds:  30,
			expected: 30,
		},
		{
			name:     "60秒",
			seconds:  60,
			expected: 60,
		},
		{
			name:     "1小时",
			seconds:  3600,
			expected: 3600,
		},
		{
			name:     "负数（非法但可测试）",
			seconds:  -1,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &DeleteOptions{}
			WithGracePeriodSeconds(tt.seconds)(opts)

			if opts.GracePeriodSeconds == nil {
				t.Fatal("GracePeriodSeconds 不应为 nil")
			}

			if *opts.GracePeriodSeconds != tt.expected {
				t.Errorf("GracePeriodSeconds = %d, want %d", *opts.GracePeriodSeconds, tt.expected)
			}
		})
	}
}

// TestForcePointerIndependence 测试 Force 指针独立性
func TestForcePointerIndependence(t *testing.T) {
	opts1 := &PatchOptions{}
	opts2 := &PatchOptions{}

	WithForce(true)(opts1)
	WithForce(false)(opts2)

	// 验证两个选项的 Force 是独立的
	if opts1.Force == opts2.Force {
		t.Error("两个 PatchOptions 的 Force 指针应该是独立的")
	}

	if *opts1.Force {
		if *opts2.Force {
			t.Error("opts2.Force 应该为 false")
		}
	} else {
		t.Error("opts1.Force 应该为 true")
	}
}
