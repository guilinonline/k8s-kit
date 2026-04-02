package resource

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// ListOptions 列表选项
type ListOptions struct {
	Namespace     string
	LabelSelector labels.Selector
	FieldSelector fields.Selector
	Limit         int64
	Continue      string
}

// ListOption 列表选项函数
type ListOption func(*ListOptions)

// WithNamespace 设置命名空间
func WithNamespace(ns string) ListOption {
	return func(o *ListOptions) {
		o.Namespace = ns
	}
}

// WithLabelSelector 设置标签选择器
func WithLabelSelector(selector labels.Selector) ListOption {
	return func(o *ListOptions) {
		o.LabelSelector = selector
	}
}

// WithLimit 设置分页限制
func WithLimit(limit int64) ListOption {
	return func(o *ListOptions) {
		o.Limit = limit
	}
}

// WithContinue 设置分页继续 token
// 用于获取下一页数据，token 来自上一次返回的 ListMeta.Continue
func WithContinue(token string) ListOption {
	return func(o *ListOptions) {
		o.Continue = token
	}
}

// CreateOptions 创建选项
type CreateOptions struct {
	FieldManager string
}

// CreateOption 创建选项函数
type CreateOption func(*CreateOptions)

// WithFieldManager 设置字段管理器
func WithFieldManager(fm string) CreateOption {
	return func(o *CreateOptions) {
		o.FieldManager = fm
	}
}

// UpdateOptions 更新选项
type UpdateOptions struct {
	FieldManager string
}

// UpdateOption 更新选项函数
type UpdateOption func(*UpdateOptions)

// WithUpdateFieldManager 设置更新字段管理器
func WithUpdateFieldManager(fm string) UpdateOption {
	return func(o *UpdateOptions) {
		o.FieldManager = fm
	}
}

// PatchOptions 补丁选项
type PatchOptions struct {
	FieldManager string
	Force        *bool
}

// PatchOption 补丁选项函数
type PatchOption func(*PatchOptions)

// WithPatchFieldManager 设置补丁字段管理器
func WithPatchFieldManager(fm string) PatchOption {
	return func(o *PatchOptions) {
		o.FieldManager = fm
	}
}

// WithForce 设置强制补丁
func WithForce(force bool) PatchOption {
	return func(o *PatchOptions) {
		o.Force = &force
	}
}

// DeleteOptions 删除选项
type DeleteOptions struct {
	GracePeriodSeconds *int64
	Preconditions      interface{} // *metav1.Preconditions
}

// DeleteOption 删除选项函数
type DeleteOption func(*DeleteOptions)

// WithGracePeriodSeconds 设置优雅期
func WithGracePeriodSeconds(seconds int64) DeleteOption {
	return func(o *DeleteOptions) {
		o.GracePeriodSeconds = &seconds
	}
}

// WatchOptions 监听选项
type WatchOptions struct {
	Namespace     string
	LabelSelector labels.Selector
	FieldSelector fields.Selector
}

// WatchOption 监听选项函数
type WatchOption func(*WatchOptions)

// WithWatchNamespace 设置监听命名空间
func WithWatchNamespace(ns string) WatchOption {
	return func(o *WatchOptions) {
		o.Namespace = ns
	}
}

// WithWatchLabelSelector 设置监听标签选择器
func WithWatchLabelSelector(selector labels.Selector) WatchOption {
	return func(o *WatchOptions) {
		o.LabelSelector = selector
	}
}
