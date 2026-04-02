package resource

import (
	"testing"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// TestWithNamespace 测试设置命名空间
func TestWithNamespace(t *testing.T) {
	opts := &ListOptions{}

	// 测试默认空值
	if opts.Namespace != "" {
		t.Errorf("默认 Namespace 应为空字符串，得到 %s", opts.Namespace)
	}

	// 应用选项
	opt := WithNamespace("default")
	opt(opts)

	if opts.Namespace != "default" {
		t.Errorf("Namespace 应为 default，得到 %s", opts.Namespace)
	}

	// 测试不同命名空间
	opt = WithNamespace("kube-system")
	opt(opts)

	if opts.Namespace != "kube-system" {
		t.Errorf("Namespace 应为 kube-system，得到 %s", opts.Namespace)
	}
}

// TestWithLabelSelector 测试设置标签选择器
func TestWithLabelSelector(t *testing.T) {
	opts := &ListOptions{}

	// 测试默认空值
	if opts.LabelSelector != nil {
		t.Error("默认 LabelSelector 应为 nil")
	}

	// 应用选项 - 使用 Everything() 选择器
	selector := labels.Everything()
	opt := WithLabelSelector(selector)
	opt(opts)

	if opts.LabelSelector.String() != selector.String() {
		t.Errorf("LabelSelector 不匹配, 得到 %v", opts.LabelSelector.String())
	}

	// 测试具体标签选择器
	selector, _ = labels.Parse("app=nginx")
	opt = WithLabelSelector(selector)
	opt(opts)

	if opts.LabelSelector.String() != selector.String() {
		t.Errorf("LabelSelector 不匹配, 得到 %v", opts.LabelSelector.String())
	}

	if opts.LabelSelector.String() != "app=nginx" {
		t.Errorf("LabelSelector 字符串应为 'app=nginx'，得到 %s", opts.LabelSelector.String())
	}
}

// TestWithLimit 测试设置分页限制
func TestWithLimit(t *testing.T) {
	opts := &ListOptions{}

	// 测试默认空值
	if opts.Limit != 0 {
		t.Errorf("默认 Limit 应为 0，得到 %d", opts.Limit)
	}

	// 应用选项
	opt := WithLimit(10)
	opt(opts)

	if opts.Limit != 10 {
		t.Errorf("Limit 应为 10，得到 %d", opts.Limit)
	}

	// 测试不同值
	opt = WithLimit(100)
	opt(opts)

	if opts.Limit != 100 {
		t.Errorf("Limit 应为 100，得到 %d", opts.Limit)
	}

	// 测试边界值
	opt = WithLimit(0)
	opt(opts)

	if opts.Limit != 0 {
		t.Errorf("Limit 应为 0，得到 %d", opts.Limit)
	}
}

// TestWithContinue 测试设置分页继续 token
func TestWithContinue(t *testing.T) {
	opts := &ListOptions{}

	// 测试默认空值
	if opts.Continue != "" {
		t.Errorf("默认 Continue 应为空字符串，得到 %s", opts.Continue)
	}

	// 应用选项
	token := "eyJjb250aW51ZVRva2VuIjoiIn0"
	opt := WithContinue(token)
	opt(opts)

	if opts.Continue != token {
		t.Errorf("Continue 应为 %s，得到 %s", token, opts.Continue)
	}

	// 测试空 token
	opt = WithContinue("")
	opt(opts)

	if opts.Continue != "" {
		t.Errorf("Continue 应为空字符串，得到 %s", opts.Continue)
	}
}

// TestWithFieldSelector 测试设置字段选择器（通过ListOptions）
func TestWithFieldSelector(t *testing.T) {
	opts := &ListOptions{}

	// 测试默认空值
	if opts.FieldSelector != nil {
		t.Error("默认 FieldSelector 应为 nil")
	}

	// 设置字段选择器
	selector := fields.Everything()
	opts.FieldSelector = selector

	if opts.FieldSelector.String() != selector.String() {
		t.Errorf("FieldSelector 不匹配, 得到 %v", opts.FieldSelector.String())
	}

	// 测试具体字段选择器
	selector, _ = fields.ParseSelector("metadata.name=my-pod")
	opts.FieldSelector = selector

	if opts.FieldSelector.String() != "metadata.name=my-pod" {
		t.Errorf("FieldSelector 字符串应为 'metadata.name=my-pod'，得到 %s", opts.FieldSelector.String())
	}
}

// TestListOptions_Combination 测试 ListOptions 组合使用
func TestListOptions_Combination(t *testing.T) {
	opts := &ListOptions{}

	// 组合多个选项
	selector, _ := labels.Parse("app=nginx,tier=frontend")

	WithNamespace("production")(opts)
	WithLabelSelector(selector)(opts)
	WithLimit(50)(opts)
	WithContinue("some-token")(opts)

	// 验证所有选项
	if opts.Namespace != "production" {
		t.Errorf("Namespace 应为 production，得到 %s", opts.Namespace)
	}

	if opts.LabelSelector == nil || opts.LabelSelector.String() != "app=nginx,tier=frontend" {
		t.Errorf("LabelSelector 不匹配，得到 %v", opts.LabelSelector)
	}

	if opts.Limit != 50 {
		t.Errorf("Limit 应为 50，得到 %d", opts.Limit)
	}

	if opts.Continue != "some-token" {
		t.Errorf("Continue 应为 some-token，得到 %s", opts.Continue)
	}
}

// TestWithFieldManager 测试设置字段管理器（CreateOptions）
func TestWithFieldManager(t *testing.T) {
	opts := &CreateOptions{}

	// 测试默认空值
	if opts.FieldManager != "" {
		t.Errorf("默认 FieldManager 应为空字符串，得到 %s", opts.FieldManager)
	}

	// 应用选项
	opt := WithFieldManager("my-controller")
	opt(opts)

	if opts.FieldManager != "my-controller" {
		t.Errorf("FieldManager 应为 my-controller，得到 %s", opts.FieldManager)
	}

	// 测试不同值
	opt = WithFieldManager("kubectl-client-side-apply")
	opt(opts)

	if opts.FieldManager != "kubectl-client-side-apply" {
		t.Errorf("FieldManager 应为 kubectl-client-side-apply，得到 %s", opts.FieldManager)
	}
}

// TestWithUpdateFieldManager 测试设置更新字段管理器（UpdateOptions）
func TestWithUpdateFieldManager(t *testing.T) {
	opts := &UpdateOptions{}

	// 测试默认空值
	if opts.FieldManager != "" {
		t.Errorf("默认 FieldManager 应为空字符串，得到 %s", opts.FieldManager)
	}

	// 应用选项
	opt := WithUpdateFieldManager("my-controller")
	opt(opts)

	if opts.FieldManager != "my-controller" {
		t.Errorf("FieldManager 应为 my-controller，得到 %s", opts.FieldManager)
	}

	// 测试不同值
	opt = WithUpdateFieldManager("another-controller")
	opt(opts)

	if opts.FieldManager != "another-controller" {
		t.Errorf("FieldManager 应为 another-controller，得到 %s", opts.FieldManager)
	}
}

// TestWithPatchFieldManager 测试设置补丁字段管理器（PatchOptions）
func TestWithPatchFieldManager(t *testing.T) {
	opts := &PatchOptions{}

	// 测试默认空值
	if opts.FieldManager != "" {
		t.Errorf("默认 FieldManager 应为空字符串，得到 %s", opts.FieldManager)
	}

	// 应用选项
	opt := WithPatchFieldManager("my-controller")
	opt(opts)

	if opts.FieldManager != "my-controller" {
		t.Errorf("FieldManager 应为 my-controller，得到 %s", opts.FieldManager)
	}

	// 测试不同值
	opt = WithPatchFieldManager("patch-controller")
	opt(opts)

	if opts.FieldManager != "patch-controller" {
		t.Errorf("FieldManager 应为 patch-controller，得到 %s", opts.FieldManager)
	}
}

// TestWithForce 测试设置强制补丁（PatchOptions）
func TestWithForce(t *testing.T) {
	opts := &PatchOptions{}

	// 测试默认空值
	if opts.Force != nil {
		t.Error("默认 Force 应为 nil")
	}

	// 应用选项 - Force = true
	opt := WithForce(true)
	opt(opts)

	if opts.Force == nil {
		t.Fatal("Force 不应为 nil")
	}
	if !*opts.Force {
		t.Error("Force 应为 true")
	}

	// 应用选项 - Force = false
	opt = WithForce(false)
	opt(opts)

	if opts.Force == nil {
		t.Fatal("Force 不应为 nil")
	}
	if *opts.Force {
		t.Error("Force 应为 false")
	}
}

// TestPatchOptions_Combination 测试 PatchOptions 组合使用
func TestPatchOptions_Combination(t *testing.T) {
	opts := &PatchOptions{}

	// 组合多个选项
	WithPatchFieldManager("my-controller")(opts)
	WithForce(true)(opts)

	// 验证所有选项
	if opts.FieldManager != "my-controller" {
		t.Errorf("FieldManager 应为 my-controller，得到 %s", opts.FieldManager)
	}

	if opts.Force == nil || !*opts.Force {
		t.Error("Force 应为 true")
	}
}

// TestWithGracePeriodSeconds 测试设置优雅期（DeleteOptions）
func TestWithGracePeriodSeconds(t *testing.T) {
	opts := &DeleteOptions{}

	// 测试默认空值
	if opts.GracePeriodSeconds != nil {
		t.Error("默认 GracePeriodSeconds 应为 nil")
	}

	// 应用选项
	opt := WithGracePeriodSeconds(30)
	opt(opts)

	if opts.GracePeriodSeconds == nil {
		t.Fatal("GracePeriodSeconds 不应为 nil")
	}
	if *opts.GracePeriodSeconds != 30 {
		t.Errorf("GracePeriodSeconds 应为 30，得到 %d", *opts.GracePeriodSeconds)
	}

	// 测试不同值 - 立即删除
	opt = WithGracePeriodSeconds(0)
	opt(opts)

	if opts.GracePeriodSeconds == nil {
		t.Fatal("GracePeriodSeconds 不应为 nil")
	}
	if *opts.GracePeriodSeconds != 0 {
		t.Errorf("GracePeriodSeconds 应为 0，得到 %d", *opts.GracePeriodSeconds)
	}

	// 测试不同值 - 60秒
	opt = WithGracePeriodSeconds(60)
	opt(opts)

	if opts.GracePeriodSeconds == nil {
		t.Fatal("GracePeriodSeconds 不应为 nil")
	}
	if *opts.GracePeriodSeconds != 60 {
		t.Errorf("GracePeriodSeconds 应为 60，得到 %d", *opts.GracePeriodSeconds)
	}
}

// TestWithWatchNamespace 测试设置监听命名空间（WatchOptions）
func TestWithWatchNamespace(t *testing.T) {
	opts := &WatchOptions{}

	// 测试默认空值
	if opts.Namespace != "" {
		t.Errorf("默认 Namespace 应为空字符串，得到 %s", opts.Namespace)
	}

	// 应用选项
	opt := WithWatchNamespace("monitoring")
	opt(opts)

	if opts.Namespace != "monitoring" {
		t.Errorf("Namespace 应为 monitoring，得到 %s", opts.Namespace)
	}

	// 测试不同值
	opt = WithWatchNamespace("ingress-nginx")
	opt(opts)

	if opts.Namespace != "ingress-nginx" {
		t.Errorf("Namespace 应为 ingress-nginx，得到 %s", opts.Namespace)
	}
}

// TestWithWatchLabelSelector 测试设置监听标签选择器（WatchOptions）
func TestWithWatchLabelSelector(t *testing.T) {
	opts := &WatchOptions{}

	// 测试默认空值
	if opts.LabelSelector != nil {
		t.Error("默认 LabelSelector 应为 nil")
	}

	// 应用选项
	selector, _ := labels.Parse("app=nginx")
	opt := WithWatchLabelSelector(selector)
	opt(opts)

	if opts.LabelSelector.String() != selector.String() {
		t.Errorf("LabelSelector 不匹配, 得到 %v", opts.LabelSelector.String())
	}

	if opts.LabelSelector.String() != "app=nginx" {
		t.Errorf("LabelSelector 字符串应为 'app=nginx'，得到 %s", opts.LabelSelector.String())
	}
}

// TestWatchOptions_Combination 测试 WatchOptions 组合使用
func TestWatchOptions_Combination(t *testing.T) {
	opts := &WatchOptions{}

	// 组合多个选项
	selector, _ := labels.Parse("app=web,tier=frontend")

	WithWatchNamespace("production")(opts)
	WithWatchLabelSelector(selector)(opts)

	// 验证所有选项
	if opts.Namespace != "production" {
		t.Errorf("Namespace 应为 production，得到 %s", opts.Namespace)
	}

	if opts.LabelSelector == nil || opts.LabelSelector.String() != "app=web,tier=frontend" {
		t.Errorf("LabelSelector 不匹配，得到 %v", opts.LabelSelector)
	}
}

// TestWatchOptions_FieldSelector 测试 WatchOptions 的 FieldSelector
func TestWatchOptions_FieldSelector(t *testing.T) {
	opts := &WatchOptions{}

	// 测试默认空值
	if opts.FieldSelector != nil {
		t.Error("默认 FieldSelector 应为 nil")
	}

	// 设置字段选择器
	selector := fields.Everything()
	opts.FieldSelector = selector

	if opts.FieldSelector.String() != selector.String() {
		t.Errorf("FieldSelector 不匹配, 得到 %v", opts.FieldSelector.String())
	}

	// 测试具体字段选择器
	selector, _ = fields.ParseSelector("metadata.namespace=default")
	opts.FieldSelector = selector

	if opts.FieldSelector.String() != "metadata.namespace=default" {
		t.Errorf("FieldSelector 字符串应为 'metadata.namespace=default'，得到 %s", opts.FieldSelector.String())
	}
}

// TestCreateOptions_Default 测试 CreateOptions 默认值
func TestCreateOptions_Default(t *testing.T) {
	opts := &CreateOptions{}

	if opts.FieldManager != "" {
		t.Errorf("FieldManager 默认应为空字符串，得到 %s", opts.FieldManager)
	}
}

// TestUpdateOptions_Default 测试 UpdateOptions 默认值
func TestUpdateOptions_Default(t *testing.T) {
	opts := &UpdateOptions{}

	if opts.FieldManager != "" {
		t.Errorf("FieldManager 默认应为空字符串，得到 %s", opts.FieldManager)
	}
}

// TestPatchOptions_Default 测试 PatchOptions 默认值
func TestPatchOptions_Default(t *testing.T) {
	opts := &PatchOptions{}

	if opts.FieldManager != "" {
		t.Errorf("FieldManager 默认应为空字符串，得到 %s", opts.FieldManager)
	}

	if opts.Force != nil {
		t.Error("Force 默认应为 nil")
	}
}

// TestDeleteOptions_Default 测试 DeleteOptions 默认值
func TestDeleteOptions_Default(t *testing.T) {
	opts := &DeleteOptions{}

	if opts.GracePeriodSeconds != nil {
		t.Error("GracePeriodSeconds 默认应为 nil")
	}

	if opts.Preconditions != nil {
		t.Error("Preconditions 默认应为 nil")
	}
}

// TestListOptions_Default 测试 ListOptions 默认值
func TestListOptions_Default(t *testing.T) {
	opts := &ListOptions{}

	if opts.Namespace != "" {
		t.Errorf("Namespace 默认应为空字符串，得到 %s", opts.Namespace)
	}

	if opts.LabelSelector != nil {
		t.Error("LabelSelector 默认应为 nil")
	}

	if opts.FieldSelector != nil {
		t.Error("FieldSelector 默认应为 nil")
	}

	if opts.Limit != 0 {
		t.Errorf("Limit 默认应为 0，得到 %d", opts.Limit)
	}

	if opts.Continue != "" {
		t.Errorf("Continue 默认应为空字符串，得到 %s", opts.Continue)
	}
}

// TestWatchOptions_Default 测试 WatchOptions 默认值
func TestWatchOptions_Default(t *testing.T) {
	opts := &WatchOptions{}

	if opts.Namespace != "" {
		t.Errorf("Namespace 默认应为空字符串，得到 %s", opts.Namespace)
	}

	if opts.LabelSelector != nil {
		t.Error("LabelSelector 默认应为 nil")
	}

	if opts.FieldSelector != nil {
		t.Error("FieldSelector 默认应为 nil")
	}
}

// TestAllOptionsTypes 测试所有选项类型的使用场景
func TestAllOptionsTypes(t *testing.T) {
	// ListOptions
	listOpts := &ListOptions{}
	WithNamespace("default")(listOpts)
	WithLimit(100)(listOpts)

	if listOpts.Namespace != "default" || listOpts.Limit != 100 {
		t.Error("ListOptions 设置失败")
	}

	// CreateOptions
	createOpts := &CreateOptions{}
	WithFieldManager("controller")(createOpts)

	if createOpts.FieldManager != "controller" {
		t.Error("CreateOptions 设置失败")
	}

	// UpdateOptions
	updateOpts := &UpdateOptions{}
	WithUpdateFieldManager("updater")(updateOpts)

	if updateOpts.FieldManager != "updater" {
		t.Error("UpdateOptions 设置失败")
	}

	// PatchOptions
	patchOpts := &PatchOptions{}
	WithPatchFieldManager("patcher")(patchOpts)
	WithForce(true)(patchOpts)

	if patchOpts.FieldManager != "patcher" || patchOpts.Force == nil || !*patchOpts.Force {
		t.Error("PatchOptions 设置失败")
	}

	// DeleteOptions
	deleteOpts := &DeleteOptions{}
	WithGracePeriodSeconds(60)(deleteOpts)

	if deleteOpts.GracePeriodSeconds == nil || *deleteOpts.GracePeriodSeconds != 60 {
		t.Error("DeleteOptions 设置失败")
	}

	// WatchOptions
	watchOpts := &WatchOptions{}
	WithWatchNamespace("kube-system")(watchOpts)

	if watchOpts.Namespace != "kube-system" {
		t.Error("WatchOptions 设置失败")
	}
}
