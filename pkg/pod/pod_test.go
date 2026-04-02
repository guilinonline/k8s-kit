package pod

import (
	"errors"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// TestNewOperator 测试 Operator 创建
func TestNewOperator(t *testing.T) {
	op := NewOperator()

	if op == nil {
		t.Fatal("NewOperator() 返回 nil")
	}
}

// TestLogOptions_Default 测试 LogOptions 默认值
func TestLogOptions_Default(t *testing.T) {
	opts := &LogOptions{}

	// 所有字段应为零值
	if opts.Container != "" {
		t.Errorf("Container 默认应为空字符串, 得到 %v", opts.Container)
	}

	if opts.Follow {
		t.Error("Follow 默认应为 false")
	}

	if opts.Previous {
		t.Error("Previous 默认应为 false")
	}

	if opts.TailLines != nil {
		t.Error("TailLines 默认应为 nil")
	}

	if opts.SinceTime != nil {
		t.Error("SinceTime 默认应为 nil")
	}

	if opts.SinceSeconds != nil {
		t.Error("SinceSeconds 默认应为 nil")
	}

	if opts.Timestamps {
		t.Error("Timestamps 默认应为 false")
	}

	if opts.LimitBytes != nil {
		t.Error("LimitBytes 默认应为 nil")
	}

	if opts.Namespace != "" {
		t.Errorf("Namespace 默认应为空字符串, 得到 %v", opts.Namespace)
	}

	if opts.LabelSelector != nil {
		t.Error("LabelSelector 默认应为 nil")
	}

	if opts.FieldSelector != nil {
		t.Error("FieldSelector 默认应为 nil")
	}
}

// TestWithContainer 测试容器选项
func TestWithContainer(t *testing.T) {
	opts := &LogOptions{}

	WithContainer("main")(opts)
	if opts.Container != "main" {
		t.Errorf("Container = %v, want main", opts.Container)
	}

	WithContainer("sidecar")(opts)
	if opts.Container != "sidecar" {
		t.Errorf("Container = %v, want sidecar", opts.Container)
	}
}

// TestWithFollow 测试跟随选项
func TestWithFollow(t *testing.T) {
	opts := &LogOptions{}

	WithFollow(true)(opts)
	if !opts.Follow {
		t.Error("Follow 应为 true")
	}

	WithFollow(false)(opts)
	if opts.Follow {
		t.Error("Follow 应为 false")
	}
}

// TestWithPrevious 测试前一个容器选项
func TestWithPrevious(t *testing.T) {
	opts := &LogOptions{}

	WithPrevious(true)(opts)
	if !opts.Previous {
		t.Error("Previous 应为 true")
	}

	WithPrevious(false)(opts)
	if opts.Previous {
		t.Error("Previous 应为 false")
	}
}

// TestWithTailLines 测试尾部行数选项
func TestWithTailLines(t *testing.T) {
	opts := &LogOptions{}

	WithTailLines(100)(opts)
	if opts.TailLines == nil || *opts.TailLines != 100 {
		t.Errorf("TailLines = %v, want 100", opts.TailLines)
	}

	WithTailLines(0)(opts)
	if opts.TailLines == nil || *opts.TailLines != 0 {
		t.Errorf("TailLines = %v, want 0", opts.TailLines)
	}

	WithTailLines(-1)(opts)
	if opts.TailLines == nil || *opts.TailLines != -1 {
		t.Errorf("TailLines = %v, want -1", opts.TailLines)
	}
}

// TestWithSinceTime 测试开始时间选项
func TestWithSinceTime(t *testing.T) {
	opts := &LogOptions{}

	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	WithSinceTime(testTime)(opts)

	if opts.SinceTime == nil {
		t.Fatal("SinceTime 不应为 nil")
	}

	if !opts.SinceTime.Equal(testTime) {
		t.Errorf("SinceTime = %v, want %v", opts.SinceTime, testTime)
	}
}

// TestWithSinceSeconds 测试持续时间选项
func TestWithSinceSeconds(t *testing.T) {
	opts := &LogOptions{}

	WithSinceSeconds(3600)(opts)
	if opts.SinceSeconds == nil || *opts.SinceSeconds != 3600 {
		t.Errorf("SinceSeconds = %v, want 3600", opts.SinceSeconds)
	}

	WithSinceSeconds(0)(opts)
	if opts.SinceSeconds == nil || *opts.SinceSeconds != 0 {
		t.Errorf("SinceSeconds = %v, want 0", opts.SinceSeconds)
	}

	WithSinceSeconds(-60)(opts)
	if opts.SinceSeconds == nil || *opts.SinceSeconds != -60 {
		t.Errorf("SinceSeconds = %v, want -60", opts.SinceSeconds)
	}
}

// TestWithTimestamps 测试时间戳选项
func TestWithTimestamps(t *testing.T) {
	opts := &LogOptions{}

	WithTimestamps(true)(opts)
	if !opts.Timestamps {
		t.Error("Timestamps 应为 true")
	}

	WithTimestamps(false)(opts)
	if opts.Timestamps {
		t.Error("Timestamps 应为 false")
	}
}

// TestWithLimitBytes 测试字节限制选项
func TestWithLimitBytes(t *testing.T) {
	opts := &LogOptions{}

	WithLimitBytes(1024)(opts)
	if opts.LimitBytes == nil || *opts.LimitBytes != 1024 {
		t.Errorf("LimitBytes = %v, want 1024", opts.LimitBytes)
	}

	WithLimitBytes(0)(opts)
	if opts.LimitBytes == nil || *opts.LimitBytes != 0 {
		t.Errorf("LimitBytes = %v, want 0", opts.LimitBytes)
	}
}

// TestWithLogNamespace 测试命名空间选项
func TestWithLogNamespace(t *testing.T) {
	opts := &LogOptions{}

	WithLogNamespace("default")(opts)
	if opts.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", opts.Namespace)
	}

	WithLogNamespace("kube-system")(opts)
	if opts.Namespace != "kube-system" {
		t.Errorf("Namespace = %v, want kube-system", opts.Namespace)
	}
}

// TestWithLogLabelSelector 测试标签选择器选项
func TestWithLogLabelSelector(t *testing.T) {
	opts := &LogOptions{}

	selector := labels.Everything()
	WithLogLabelSelector(selector)(opts)
	if opts.LabelSelector.String() != selector.String() {
		t.Errorf("LabelSelector = %v, want %v", opts.LabelSelector.String(), selector.String())
	}

	selector, _ = labels.Parse("app=nginx")
	WithLogLabelSelector(selector)(opts)
	if opts.LabelSelector.String() != selector.String() {
		t.Errorf("LabelSelector = %v, want %v", opts.LabelSelector.String(), selector.String())
	}
}

// TestWithLogFieldSelector 测试字段选择器选项
func TestWithLogFieldSelector(t *testing.T) {
	opts := &LogOptions{}

	selector := fields.Everything()
	WithLogFieldSelector(selector)(opts)
	if opts.FieldSelector.String() != selector.String() {
		t.Errorf("FieldSelector = %v, want %v", opts.FieldSelector.String(), selector.String())
	}

	selector, _ = fields.ParseSelector("metadata.name=my-pod")
	WithLogFieldSelector(selector)(opts)
	if opts.FieldSelector.String() != selector.String() {
		t.Errorf("FieldSelector = %v, want %v", opts.FieldSelector.String(), selector.String())
	}
}

// TestLogOptions_Combination 测试 LogOptions 组合
func TestLogOptions_Combination(t *testing.T) {
	opts := &LogOptions{}

	// 组合多个选项
	logOpts := []LogOption{
		WithContainer("app"),
		WithFollow(true),
		WithTailLines(100),
		WithTimestamps(true),
		WithLogNamespace("production"),
	}

	for _, opt := range logOpts {
		opt(opts)
	}

	// 验证所有选项
	if opts.Container != "app" {
		t.Errorf("Container = %v, want app", opts.Container)
	}

	if !opts.Follow {
		t.Error("Follow 应为 true")
	}

	if opts.TailLines == nil || *opts.TailLines != 100 {
		t.Errorf("TailLines = %v, want 100", opts.TailLines)
	}

	if !opts.Timestamps {
		t.Error("Timestamps 应为 true")
	}

	if opts.Namespace != "production" {
		t.Errorf("Namespace = %v, want production", opts.Namespace)
	}
}

// TestExecOptions_Default 测试 ExecOptions 默认值
func TestExecOptions_Default(t *testing.T) {
	opts := &ExecOptions{}

	if opts.Container != "" {
		t.Errorf("Container 默认应为空字符串, 得到 %v", opts.Container)
	}

	if opts.TTY {
		t.Error("TTY 默认应为 false")
	}

	if opts.Stdin {
		t.Error("Stdin 默认应为 false")
	}

	if opts.Timeout != 0 {
		t.Errorf("Timeout 默认应为 0, 得到 %v", opts.Timeout)
	}
}

// TestWithExecContainer 测试执行容器选项
func TestWithExecContainer(t *testing.T) {
	opts := &ExecOptions{}

	WithExecContainer("main")(opts)
	if opts.Container != "main" {
		t.Errorf("Container = %v, want main", opts.Container)
	}
}

// TestWithTTY 测试 TTY 选项
func TestWithTTY(t *testing.T) {
	opts := &ExecOptions{}

	WithTTY(true)(opts)
	if !opts.TTY {
		t.Error("TTY 应为 true")
	}

	WithTTY(false)(opts)
	if opts.TTY {
		t.Error("TTY 应为 false")
	}
}

// TestWithStdin 测试 Stdin 选项
func TestWithStdin(t *testing.T) {
	opts := &ExecOptions{}

	WithStdin(true)(opts)
	if !opts.Stdin {
		t.Error("Stdin 应为 true")
	}

	WithStdin(false)(opts)
	if opts.Stdin {
		t.Error("Stdin 应为 false")
	}
}

// TestWithExecTimeout 测试执行超时选项
func TestWithExecTimeout(t *testing.T) {
	opts := &ExecOptions{}

	WithExecTimeout(30 * time.Second)(opts)
	if opts.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", opts.Timeout)
	}

	WithExecTimeout(0)(opts)
	if opts.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0", opts.Timeout)
	}

	WithExecTimeout(-1 * time.Second)(opts)
	if opts.Timeout != -1*time.Second {
		t.Errorf("Timeout = %v, want -1s", opts.Timeout)
	}
}

// TestExecOptions_Combination 测试 ExecOptions 组合
func TestExecOptions_Combination(t *testing.T) {
	opts := &ExecOptions{}

	// 组合多个选项
	execOpts := []ExecOption{
		WithExecContainer("app"),
		WithTTY(true),
		WithStdin(true),
		WithExecTimeout(60 * time.Second),
	}

	for _, opt := range execOpts {
		opt(opts)
	}

	// 验证所有选项
	if opts.Container != "app" {
		t.Errorf("Container = %v, want app", opts.Container)
	}

	if !opts.TTY {
		t.Error("TTY 应为 true")
	}

	if !opts.Stdin {
		t.Error("Stdin 应为 true")
	}

	if opts.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", opts.Timeout)
	}
}

// TestExecResult_Struct 测试 ExecResult 结构
func TestExecResult_Struct(t *testing.T) {
	result := ExecResult{
		Stdout:   "Hello World",
		Stderr:   "Error message",
		ExitCode: 0,
	}

	if result.Stdout != "Hello World" {
		t.Errorf("Stdout = %v, want 'Hello World'", result.Stdout)
	}

	if result.Stderr != "Error message" {
		t.Errorf("Stderr = %v, want 'Error message'", result.Stderr)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", result.ExitCode)
	}
}

// TestExecResult_ExitCodes 测试不同退出码
func TestExecResult_ExitCodes(t *testing.T) {
	tests := []struct {
		exitCode int
		expected int
	}{
		{0, 0},
		{1, 1},
		{127, 127},
		{-1, -1},
		{255, 255},
	}

	for _, tt := range tests {
		result := ExecResult{ExitCode: tt.exitCode}
		if result.ExitCode != tt.expected {
			t.Errorf("ExitCode = %v, want %v", result.ExitCode, tt.expected)
		}
	}
}

// TestIsContainerNotFound 测试容器未找到错误检查
func TestIsContainerNotFound(t *testing.T) {
	// 测试匹配的错误
	err := errors.New("container \"main\" not found")
	if !IsContainerNotFound(err) {
		t.Error("应该匹配 container not found 错误")
	}

	// 测试不匹配的错误
	err = errors.New("pod not found")
	if IsContainerNotFound(err) {
		t.Error("不应匹配 pod not found 错误")
	}

	// 测试 nil 错误
	if IsContainerNotFound(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestIsPodNotFound 测试 Pod 未找到错误检查
func TestIsPodNotFound(t *testing.T) {
	// 测试匹配的错误
	err := errors.New("pods \"my-pod\" not found")
	if !IsPodNotFound(err) {
		t.Error("应该匹配 pods not found 错误")
	}

	// 测试不匹配的错误
	err = errors.New("container not found")
	if IsPodNotFound(err) {
		t.Error("不应匹配 container not found 错误")
	}

	// 测试 nil 错误
	if IsPodNotFound(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestIsForbidden 测试权限错误检查
func TestIsForbidden(t *testing.T) {
	// 测试匹配的错误
	err := errors.New("Forbidden: User cannot access resource")
	if !IsForbidden(err) {
		t.Error("应该匹配 Forbidden 错误")
	}

	err = errors.New("HTTP 403: access denied")
	if !IsForbidden(err) {
		t.Error("应该匹配 403 错误")
	}

	// 测试不匹配的错误
	err = errors.New("connection refused")
	if IsForbidden(err) {
		t.Error("不应匹配 connection refused 错误")
	}

	// 测试 nil 错误
	if IsForbidden(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestIsTimeout 测试超时错误检查
func TestIsTimeout(t *testing.T) {
	// 测试匹配的错误
	err := errors.New("operation timeout")
	if !IsTimeout(err) {
		t.Error("应该匹配 timeout 错误")
	}

	err = errors.New("context deadline exceeded")
	if !IsTimeout(err) {
		t.Error("应该匹配 context deadline exceeded 错误")
	}

	// 测试不匹配的错误
	err = errors.New("connection refused")
	if IsTimeout(err) {
		t.Error("不应匹配 connection refused 错误")
	}

	// 测试 nil 错误
	if IsTimeout(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestIsConnectionLost 测试连接丢失错误检查
func TestIsConnectionLost(t *testing.T) {
	// 测试匹配的错误
	err := errors.New("connection lost")
	if !IsConnectionLost(err) {
		t.Error("应该匹配 connection lost 错误")
	}

	err = errors.New("connection broken")
	if !IsConnectionLost(err) {
		t.Error("应该匹配 connection broken 错误")
	}

	err = errors.New("broken connection")
	if !IsConnectionLost(err) {
		t.Error("应该匹配 broken connection 错误")
	}

	// 测试不匹配的错误
	err = errors.New("timeout")
	if IsConnectionLost(err) {
		t.Error("不应匹配 timeout 错误")
	}

	// 测试 nil 错误
	if IsConnectionLost(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestIsNotFound 测试未找到错误检查
func TestIsNotFound(t *testing.T) {
	// 测试 Pod not found
	err := errors.New("pods \"my-pod\" not found")
	if !IsNotFound(err) {
		t.Error("应该匹配 pods not found 错误")
	}

	// 测试 Container not found
	err = errors.New("container \"main\" not found")
	if !IsNotFound(err) {
		t.Error("应该匹配 container not found 错误")
	}

	// 测试不匹配的错误
	err = errors.New("timeout")
	if IsNotFound(err) {
		t.Error("不应匹配 timeout 错误")
	}

	// 测试 nil 错误
	if IsNotFound(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestIsServerError 测试服务器错误检查
func TestIsServerError(t *testing.T) {
	// 测试匹配的错误
	tests := []string{
		"HTTP 500: internal server error",
		"HTTP 502: bad gateway",
		"HTTP 503: service unavailable",
		"HTTP 504: gateway timeout",
	}

	for _, msg := range tests {
		err := errors.New(msg)
		if !IsServerError(err) {
			t.Errorf("应该匹配服务器错误: %s", msg)
		}
	}

	// 测试不匹配的错误
	err := errors.New("HTTP 404: not found")
	if IsServerError(err) {
		t.Error("不应匹配 404 错误")
	}

	// 测试 nil 错误
	if IsServerError(nil) {
		t.Error("nil 错误不应匹配")
	}
}

// TestErrorConstants 测试错误常量
func TestErrorConstants(t *testing.T) {
	// 验证错误消息
	if ErrContainerNotFound == nil {
		t.Error("ErrContainerNotFound 不应为 nil")
	}

	if ErrPodNotFound == nil {
		t.Error("ErrPodNotFound 不应为 nil")
	}

	if ErrForbidden == nil {
		t.Error("ErrForbidden 不应为 nil")
	}

	if ErrTimeout == nil {
		t.Error("ErrTimeout 不应为 nil")
	}

	if ErrConnectionLost == nil {
		t.Error("ErrConnectionLost 不应为 nil")
	}

	// 验证错误消息内容
	if ErrContainerNotFound.Error() != "container not found in pod" {
		t.Errorf("ErrContainerNotFound 消息不匹配: %v", ErrContainerNotFound.Error())
	}

	if ErrPodNotFound.Error() != "pod not found" {
		t.Errorf("ErrPodNotFound 消息不匹配: %v", ErrPodNotFound.Error())
	}
}

// BenchmarkLogOptionsApply LogOptions 应用基准测试
func BenchmarkLogOptionsApply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		opts := &LogOptions{}
		WithContainer("app")(opts)
		WithFollow(true)(opts)
		WithTailLines(100)(opts)
	}
}

// BenchmarkExecOptionsApply ExecOptions 应用基准测试
func BenchmarkExecOptionsApply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		opts := &ExecOptions{}
		WithExecContainer("app")(opts)
		WithTTY(true)(opts)
		WithExecTimeout(30 * time.Second)(opts)
	}
}
