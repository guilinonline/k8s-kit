package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/guilinonline/k8s-kit/pkg/client"
)

// Operator 资源操作器
type Operator struct{}

// NewOperator 创建操作器
func NewOperator() *Operator {
	return &Operator{}
}

// Get 获取资源
func (o *Operator) Get(
	ctx context.Context,
	cli *client.ClusterClient,
	obj client.Object,
	key types.NamespacedName,
) error {
	return cli.RuntimeClient.Get(ctx, key, obj)
}

// List 列表查询
func (o *Operator) List(
	ctx context.Context,
	cli *client.ClusterClient,
	list client.ObjectList,
	opts ...ListOption,
) error {
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	listOpts := []client.ListOption{}
	if options.Namespace != "" {
		listOpts = append(listOpts, client.InNamespace(options.Namespace))
	}
	if options.LabelSelector != nil {
		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: options.LabelSelector})
	}
	if options.FieldSelector != nil {
		listOpts = append(listOpts, client.MatchingFieldsSelector{Selector: options.FieldSelector})
	}
	if options.Limit > 0 {
		listOpts = append(listOpts, client.Limit(options.Limit))
	}
	if options.Continue != "" {
		listOpts = append(listOpts, client.Continue(options.Continue))
	}

	return cli.RuntimeClient.List(ctx, list, listOpts...)
}

// Create 创建资源
func (o *Operator) Create(
	ctx context.Context,
	cli *client.ClusterClient,
	obj client.Object,
	opts ...CreateOption,
) error {
	options := &CreateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	createOpts := []client.CreateOption{}
	if options.FieldManager != "" {
		createOpts = append(createOpts, client.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Create(ctx, obj, createOpts...)
}

// Update 更新资源
func (o *Operator) Update(
	ctx context.Context,
	cli *client.ClusterClient,
	obj client.Object,
	opts ...UpdateOption,
) error {
	options := &UpdateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	updateOpts := []client.UpdateOption{}
	if options.FieldManager != "" {
		updateOpts = append(updateOpts, client.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Update(ctx, obj, updateOpts...)
}

// Patch 补丁更新
func (o *Operator) Patch(
	ctx context.Context,
	cli *client.ClusterClient,
	obj client.Object,
	patch client.Patch,
	opts ...PatchOption,
) error {
	options := &PatchOptions{}
	for _, opt := range opts {
		opt(options)
	}

	patchOpts := []client.PatchOption{}
	if options.FieldManager != "" {
		patchOpts = append(patchOpts, client.FieldOwner(options.FieldManager))
	}
	if options.Force != nil {
		patchOpts = append(patchOpts, client.Force(*options.Force))
	}

	return cli.RuntimeClient.Patch(ctx, obj, patch, patchOpts...)
}

// Delete 删除资源
func (o *Operator) Delete(
	ctx context.Context,
	cli *client.ClusterClient,
	obj client.Object,
	opts ...DeleteOption,
) error {
	options := &DeleteOptions{}
	for _, opt := range opts {
		opt(options)
	}

	deleteOpts := []client.DeleteOption{}
	if options.GracePeriodSeconds != nil {
		deleteOpts = append(deleteOpts, client.GracePeriodSeconds(*options.GracePeriodSeconds))
	}
	if options.Preconditions != nil {
		deleteOpts = append(deleteOpts, client.Preconditions(*options.Preconditions))
	}

	return cli.RuntimeClient.Delete(ctx, obj, deleteOpts...)
}

// Watch 监听资源变化
func (o *Operator) Watch(
	ctx context.Context,
	cli *client.ClusterClient,
	obj client.Object,
	opts ...WatchOption,
) (watch.Interface, error) {
	options := &WatchOptions{}
	for _, opt := range opts {
		opt(options)
	}

	watchOpts := []client.ListOption{}
	if options.Namespace != "" {
		watchOpts = append(watchOpts, client.InNamespace(options.Namespace))
	}
	if options.LabelSelector != nil {
		watchOpts = append(watchOpts, client.MatchingLabelsSelector{Selector: options.LabelSelector})
	}
	if options.FieldSelector != nil {
		watchOpts = append(watchOpts, client.MatchingFieldsSelector{Selector: options.FieldSelector})
	}

	return cli.RuntimeClient.Watch(ctx, obj, watchOpts...)
}
