package resource

import (
	"regexp"
	"strings"

	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Object 是资源对象的接口约束
type Object interface {
	runtimeclient.Object
}

// MatchType 定义名称匹配类型
type MatchType int

const (
	// MatchContains 包含匹配（默认）
	MatchContains MatchType = iota
	// MatchPrefix 前缀匹配
	MatchPrefix
	// MatchSuffix 后缀匹配
	MatchSuffix
	// MatchRegex 正则匹配
	MatchRegex
)

// FilterByName 按名称过滤资源（客户端内存过滤）
//
// ⚠️ 重要提示：此函数在客户端内存中过滤，适合以下场景：
//  1. 已知数据量较小（< 1000 条，或配合 LabelSelector 使用后）
//  2. 不分页展示（如导出、批量操作、统计）
//  3. 不在分页列表中使用（避免"一页只有几条"的问题）
//
// 如果数据量大，建议：
//   - 先用 WithLabelSelector 减少数据量，再用此函数过滤
//   - 或使用 Informer 缓存后查询
//   - 或资源创建时将名称也作为标签
//
// 支持：包含、前缀、后缀、正则匹配
//
// 示例：
//
//	// 配合 LabelSelector 使用（推荐）
//	operator.List(ctx, cli, podList, resource.WithLabelSelector(selector))
//	filtered := resource.FilterByName(podList.Items, "nginx")
//
//	// 直接过滤
//	filtered := resource.FilterByName(items, "nginx")                    // 包含 "nginx"
//	filtered := resource.FilterByName(items, "web-", resource.MatchPrefix)  // 前缀 "web-"
//	filtered := resource.FilterByName(items, "-prod", resource.MatchSuffix) // 后缀 "-prod"
//	filtered := resource.FilterByName(items, "^web-.*-prod$", resource.MatchRegex) // 正则匹配
func FilterByName[T Object](items []T, pattern string, matchType ...MatchType) []T {
	if pattern == "" {
		return items
	}

	// 默认使用包含匹配
	mt := MatchContains
	if len(matchType) > 0 {
		mt = matchType[0]
	}

	var filtered []T
	for _, item := range items {
		name := item.GetName()
		matched := false

		switch mt {
		case MatchContains:
			matched = strings.Contains(name, pattern)
		case MatchPrefix:
			matched = strings.HasPrefix(name, pattern)
		case MatchSuffix:
			matched = strings.HasSuffix(name, pattern)
		case MatchRegex:
			matched = regexp.MustCompile(pattern).MatchString(name)
		}

		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
