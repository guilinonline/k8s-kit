package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	k8sclient "github.com/guilinonline/k8s-kit/pkg/client"
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
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	key types.NamespacedName,
) error {
	return cli.RuntimeClient.Get(ctx, key, obj)
}

// List 列表查询
func (o *Operator) List(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	list runtimeclient.ObjectList,
	opts ...ListOption,
) error {
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	listOpts := []runtimeclient.ListOption{}
	if options.Namespace != "" {
		listOpts = append(listOpts, runtimeclient.InNamespace(options.Namespace))
	}
	if options.LabelSelector != nil {
		listOpts = append(listOpts, runtimeclient.MatchingLabelsSelector{Selector: options.LabelSelector})
	}
	if options.FieldSelector != nil {
		listOpts = append(listOpts, runtimeclient.MatchingFieldsSelector{Selector: options.FieldSelector})
	}
	if options.Limit > 0 {
		listOpts = append(listOpts, runtimeclient.Limit(options.Limit))
	}
	if options.Continue != "" {
		listOpts = append(listOpts, runtimeclient.Continue(options.Continue))
	}

	return cli.RuntimeClient.List(ctx, list, listOpts...)
}

// Create 创建资源
func (o *Operator) Create(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	opts ...CreateOption,
) error {
	options := &CreateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	createOpts := []runtimeclient.CreateOption{}
	if options.FieldManager != "" {
		createOpts = append(createOpts, runtimeclient.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Create(ctx, obj, createOpts...)
}

// Update 更新资源
func (o *Operator) Update(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	opts ...UpdateOption,
) error {
	options := &UpdateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	updateOpts := []runtimeclient.UpdateOption{}
	if options.FieldManager != "" {
		updateOpts = append(updateOpts, runtimeclient.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Update(ctx, obj, updateOpts...)
}

// Patch 补丁更新
func (o *Operator) Patch(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	patch runtimeclient.Patch,
	opts ...PatchOption,
) error {
	options := &PatchOptions{}
	for _, opt := range opts {
		opt(options)
	}

	patchOpts := []runtimeclient.PatchOption{}
	if options.FieldManager != "" {
		patchOpts = append(patchOpts, runtimeclient.FieldOwner(options.FieldManager))
	}

	return cli.RuntimeClient.Patch(ctx, obj, patch, patchOpts...)
}

// Delete 删除资源
func (o *Operator) Delete(
	ctx context.Context,
	cli *k8sclient.ClusterClient,
	obj runtimeclient.Object,
	opts ...DeleteOption,
) error {
	options := &DeleteOptions{}
	for _, opt := range opts {
		opt(options)
	}

	deleteOpts := []runtimeclient.DeleteOption{}
	if options.GracePeriodSeconds != nil {
		deleteOpts = append(deleteOpts, runtimeclient.GracePeriodSeconds(*options.GracePeriodSeconds))
	}

	return cli.RuntimeClient.Delete(ctx, obj, deleteOpts...)
}
