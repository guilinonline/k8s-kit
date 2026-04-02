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

// FilterByName 按名称过滤资源
// 支持：包含、前缀、后缀、正则匹配
//
// 示例：
//
//	resource.FilterByName(items, "nginx")                    // 包含 "nginx"
//	resource.FilterByName(items, "web-", resource.MatchPrefix)  // 前缀 "web-"
//	resource.FilterByName(items, "-prod", resource.MatchSuffix) // 后缀 "-prod"
//	resource.FilterByName(items, "^web-.*-prod$", resource.MatchRegex) // 正则匹配
//
// 注意：此函数在客户端内存中过滤，适合数据量较小的场景
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
