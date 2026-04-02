package resource

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createTestPods 创建测试用的 Pod 列表（指针类型）
func createTestPods() []*corev1.Pod {
	return []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "nginx-web-1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "nginx-web-2"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "nginx-api-1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "redis-cache-1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "redis-cache-2"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "web-frontend"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "web-backend"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "api-gateway"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "mysql-prod-1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "mysql-staging-2"}},
	}
}

// TestFilterByName_Contains 测试包含匹配（默认）
func TestFilterByName_Contains(t *testing.T) {
	pods := createTestPods()

	tests := []struct {
		name     string
		pattern  string
		expected int
		contains []string
	}{
		{
			name:     "匹配所有包含 nginx 的 Pod",
			pattern:  "nginx",
			expected: 3,
			contains: []string{"nginx-web-1", "nginx-web-2", "nginx-api-1"},
		},
		{
			name:     "匹配所有包含 redis 的 Pod",
			pattern:  "redis",
			expected: 2,
			contains: []string{"redis-cache-1", "redis-cache-2"},
		},
		{
			name:     "匹配所有包含 web 的 Pod",
			pattern:  "web",
			expected: 4,
			contains: []string{"nginx-web-1", "nginx-web-2", "web-frontend", "web-backend"},
		},
		{
			name:     "匹配所有包含 mysql 的 Pod",
			pattern:  "mysql",
			expected: 2,
			contains: []string{"mysql-prod-1", "mysql-staging-2"},
		},
		{
			name:     "匹配所有包含 api 的 Pod",
			pattern:  "api",
			expected: 2,
			contains: []string{"nginx-api-1", "api-gateway"},
		},
		{
			name:     "匹配不存在的模式",
			pattern:  "notfound",
			expected: 0,
			contains: []string{},
		},
		{
			name:     "空模式应返回所有",
			pattern:  "",
			expected: 10,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByName(pods, tt.pattern)
			if len(result) != tt.expected {
				t.Errorf("期望 %d 个结果，得到 %d 个", tt.expected, len(result))
			}
			for _, name := range tt.contains {
				found := false
				for _, pod := range result {
					if pod.Name == name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("期望结果中包含 %s，但没有找到", name)
				}
			}
		})
	}
}

// TestFilterByName_Prefix 测试前缀匹配
func TestFilterByName_Prefix(t *testing.T) {
	pods := createTestPods()

	tests := []struct {
		name     string
		pattern  string
		expected int
		contains []string
	}{
		{
			name:     "前缀匹配 nginx",
			pattern:  "nginx",
			expected: 3,
			contains: []string{"nginx-web-1", "nginx-web-2", "nginx-api-1"},
		},
		{
			name:     "前缀匹配 redis",
			pattern:  "redis",
			expected: 2,
			contains: []string{"redis-cache-1", "redis-cache-2"},
		},
		{
			name:     "前缀匹配 web",
			pattern:  "web",
			expected: 2,
			contains: []string{"web-frontend", "web-backend"},
		},
		{
			name:     "前缀匹配 api",
			pattern:  "api",
			expected: 1,
			contains: []string{"api-gateway"},
		},
		{
			name:     "前缀不匹配任何项",
			pattern:  "web-xxx",
			expected: 0,
			contains: []string{},
		},
		{
			name:     "空前缀应返回所有",
			pattern:  "",
			expected: 10,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByName(pods, tt.pattern, MatchPrefix)
			if len(result) != tt.expected {
				t.Errorf("期望 %d 个结果，得到 %d 个", tt.expected, len(result))
			}
			for _, name := range tt.contains {
				found := false
				for _, pod := range result {
					if pod.Name == name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("期望结果中包含 %s，但没有找到", name)
				}
			}
		})
	}
}

// TestFilterByName_Suffix 测试后缀匹配
func TestFilterByName_Suffix(t *testing.T) {
	pods := createTestPods()

	tests := []struct {
		name     string
		pattern  string
		expected int
		contains []string
	}{
		{
			name:     "后缀匹配 -1",
			pattern:  "-1",
			expected: 4,
			contains: []string{"nginx-web-1", "nginx-api-1", "redis-cache-1", "mysql-prod-1"},
		},
		{
			name:     "后缀匹配 -2",
			pattern:  "-2",
			expected: 3,
			contains: []string{"nginx-web-2", "redis-cache-2", "mysql-staging-2"},
		},
		{
			name:     "后缀匹配 -prod-1",
			pattern:  "-prod-1",
			expected: 1,
			contains: []string{"mysql-prod-1"},
		},
		{
			name:     "后缀匹配 -staging-2",
			pattern:  "-staging-2",
			expected: 1,
			contains: []string{"mysql-staging-2"},
		},
		{
			name:     "后缀不匹配任何项",
			pattern:  "-notfound",
			expected: 0,
			contains: []string{},
		},
		{
			name:     "空后缀应返回所有",
			pattern:  "",
			expected: 10,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByName(pods, tt.pattern, MatchSuffix)
			if len(result) != tt.expected {
				t.Errorf("期望 %d 个结果，得到 %d 个", tt.expected, len(result))
			}
			for _, name := range tt.contains {
				found := false
				for _, pod := range result {
					if pod.Name == name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("期望结果中包含 %s，但没有找到", name)
				}
			}
		})
	}
}

// TestFilterByName_Regex 测试正则匹配
func TestFilterByName_Regex(t *testing.T) {
	pods := createTestPods()

	tests := []struct {
		name     string
		pattern  string
		expected int
		contains []string
	}{
		{
			name:     "正则匹配 nginx-web-.*",
			pattern:  "^nginx-web-.*",
			expected: 2,
			contains: []string{"nginx-web-1", "nginx-web-2"},
		},
		{
			name:     "正则匹配 .*-cache-.*",
			pattern:  ".*-cache-.*",
			expected: 2,
			contains: []string{"redis-cache-1", "redis-cache-2"},
		},
		{
			name:     "正则匹配 web-.*",
			pattern:  "^web-.*",
			expected: 2,
			contains: []string{"web-frontend", "web-backend"},
		},
		{
			name:     "正则匹配 mysql-.*",
			pattern:  "^mysql-.*",
			expected: 2,
			contains: []string{"mysql-prod-1", "mysql-staging-2"},
		},
		{
			name:     "正则匹配 .*-prod-.*",
			pattern:  ".*-prod-.*",
			expected: 1,
			contains: []string{"mysql-prod-1"},
		},
		{
			name:     "正则匹配不匹配任何项",
			pattern:  "^xyz-.*",
			expected: 0,
			contains: []string{},
		},
		{
			name:     "空正则应返回所有",
			pattern:  "",
			expected: 10,
			contains: []string{},
		},
		{
			name:     "复杂正则匹配",
			pattern:  "^(nginx|redis)-.*",
			expected: 5,
			contains: []string{"nginx-web-1", "nginx-web-2", "nginx-api-1", "redis-cache-1", "redis-cache-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByName(pods, tt.pattern, MatchRegex)
			if len(result) != tt.expected {
				t.Errorf("期望 %d 个结果，得到 %d 个", tt.expected, len(result))
			}
			for _, name := range tt.contains {
				found := false
				for _, pod := range result {
					if pod.Name == name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("期望结果中包含 %s，但没有找到", name)
				}
			}
		})
	}
}

// TestFilterByName_EmptySlice 测试空切片
func TestFilterByName_EmptySlice(t *testing.T) {
	emptyPods := []*corev1.Pod{}

	result := FilterByName(emptyPods, "nginx")
	if len(result) != 0 {
		t.Errorf("空切片应返回空结果，得到 %d 个", len(result))
	}

	// 各种匹配类型都应处理空切片
	result = FilterByName(emptyPods, "nginx", MatchPrefix)
	if len(result) != 0 {
		t.Errorf("空切片 + Prefix 应返回空结果，得到 %d 个", len(result))
	}

	result = FilterByName(emptyPods, "nginx", MatchSuffix)
	if len(result) != 0 {
		t.Errorf("空切片 + Suffix 应返回空结果，得到 %d 个", len(result))
	}

	result = FilterByName(emptyPods, "nginx", MatchRegex)
	if len(result) != 0 {
		t.Errorf("空切片 + Regex 应返回空结果，得到 %d 个", len(result))
	}
}

// TestFilterByName_InvalidRegex 测试无效正则
func TestFilterByName_InvalidRegex(t *testing.T) {
	pods := createTestPods()

	// 无效正则应该 panic，这是预期的行为
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("无效正则应触发 panic")
		}
	}()

	FilterByName(pods, "[invalid(", MatchRegex)
}

// TestPaginate_Normal 测试正常分页
func TestPaginate_Normal(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		name         string
		limit        int
		offset       int
		expectedLen  int
		expectedVals []int
		expectedMore bool
	}{
		{
			name:         "第一页，limit=3",
			limit:        3,
			offset:       0,
			expectedLen:  3,
			expectedVals: []int{1, 2, 3},
			expectedMore: true,
		},
		{
			name:         "第二页，limit=3",
			limit:        3,
			offset:       3,
			expectedLen:  3,
			expectedVals: []int{4, 5, 6},
			expectedMore: true,
		},
		{
			name:         "第三页，limit=3",
			limit:        3,
			offset:       6,
			expectedLen:  3,
			expectedVals: []int{7, 8, 9},
			expectedMore: true,
		},
		{
			name:         "第四页，limit=3（最后一页）",
			limit:        3,
			offset:       9,
			expectedLen:  1,
			expectedVals: []int{10},
			expectedMore: false,
		},
		{
			name:         "limit=5",
			limit:        5,
			offset:       0,
			expectedLen:  5,
			expectedVals: []int{1, 2, 3, 4, 5},
			expectedMore: true,
		},
		{
			name:         "limit=10（正好全部）",
			limit:        10,
			offset:       0,
			expectedLen:  10,
			expectedVals: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expectedMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, hasMore := Paginate(items, tt.limit, tt.offset)
			if len(result) != tt.expectedLen {
				t.Errorf("期望 %d 个结果，得到 %d 个", tt.expectedLen, len(result))
			}
			if hasMore != tt.expectedMore {
				t.Errorf("期望 hasMore=%v，得到 %v", tt.expectedMore, hasMore)
			}
			for i, v := range tt.expectedVals {
				if i >= len(result) || result[i] != v {
					t.Errorf("期望 result[%d]=%d", i, v)
				}
			}
		})
	}
}

// TestPaginate_EdgeCases 测试分页边界情况
func TestPaginate_EdgeCases(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	tests := []struct {
		name         string
		limit        int
		offset       int
		expectedLen  int
		expectedMore bool
	}{
		{
			name:         "limit=0 应返回全部",
			limit:        0,
			offset:       0,
			expectedLen:  5,
			expectedMore: false,
		},
		{
			name:         "limit=-1 应返回全部",
			limit:        -1,
			offset:       0,
			expectedLen:  5,
			expectedMore: false,
		},
		{
			name:         "offset 超过长度应返回空",
			limit:        3,
			offset:       10,
			expectedLen:  0,
			expectedMore: false,
		},
		{
			name:         "offset 等于长度应返回空",
			limit:        3,
			offset:       5,
			expectedLen:  0,
			expectedMore: false,
		},
		{
			name:         "offset 为负",
			limit:        3,
			offset:       -5,
			expectedLen:  3,
			expectedMore: true,
		},
		{
			name:         "空切片",
			limit:        3,
			offset:       0,
			expectedLen:  0,
			expectedMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testItems := items
			if tt.name == "空切片" {
				testItems = []int{}
			}
			result, hasMore := Paginate(testItems, tt.limit, tt.offset)
			if len(result) != tt.expectedLen {
				t.Errorf("期望 %d 个结果，得到 %d 个", tt.expectedLen, len(result))
			}
			if hasMore != tt.expectedMore {
				t.Errorf("期望 hasMore=%v，得到 %v", tt.expectedMore, hasMore)
			}
		})
	}
}

// TestPaginate_WithPods 测试用 Pod 进行分页
func TestPaginate_WithPods(t *testing.T) {
	pods := createTestPods()

	// 第一页，每页3个
	page1, hasMore1 := Paginate(pods, 3, 0)
	if len(page1) != 3 {
		t.Errorf("第一页期望 3 个 Pod，得到 %d 个", len(page1))
	}
	if !hasMore1 {
		t.Error("第一页期望 hasMore=true")
	}

	// 第二页
	page2, hasMore2 := Paginate(pods, 3, 3)
	if len(page2) != 3 {
		t.Errorf("第二页期望 3 个 Pod，得到 %d 个", len(page2))
	}
	if !hasMore2 {
		t.Error("第二页期望 hasMore=true")
	}

	// 第三页
	page3, hasMore3 := Paginate(pods, 3, 6)
	if len(page3) != 3 {
		t.Errorf("第三页期望 3 个 Pod，得到 %d 个", len(page3))
	}
	if !hasMore3 {
		t.Error("第三页期望 hasMore=true")
	}

	// 第四页（最后一页）
	page4, hasMore4 := Paginate(pods, 3, 9)
	if len(page4) != 1 {
		t.Errorf("第四页期望 1 个 Pod，得到 %d 个", len(page4))
	}
	if hasMore4 {
		t.Error("第四页期望 hasMore=false")
	}
}

// TestFilterAndPaginate 测试过滤后分页的完整流程
func TestFilterAndPaginate(t *testing.T) {
	pods := createTestPods()

	// 场景1：过滤包含 "web" 的 Pod，然后分页
	filtered := FilterByName(pods, "web")
	if len(filtered) != 4 {
		t.Fatalf("期望过滤后 4 个 Pod，得到 %d 个", len(filtered))
	}

	page1, hasMore := Paginate(filtered, 2, 0)
	if len(page1) != 2 {
		t.Errorf("第一页期望 2 个 Pod，得到 %d 个", len(page1))
	}
	if !hasMore {
		t.Error("期望还有更多数据")
	}

	page2, hasMore2 := Paginate(filtered, 2, 2)
	if len(page2) != 2 {
		t.Errorf("第二页期望 2 个 Pod，得到 %d 个", len(page2))
	}
	if hasMore2 {
		t.Error("期望没有更多数据")
	}

	// 场景2：前缀过滤后分页
	filtered = FilterByName(pods, "nginx", MatchPrefix)
	if len(filtered) != 3 {
		t.Fatalf("期望过滤后 3 个 Pod，得到 %d 个", len(filtered))
	}

	page1, hasMore = Paginate(filtered, 2, 0)
	if len(page1) != 2 {
		t.Errorf("第一页期望 2 个 Pod，得到 %d 个", len(page1))
	}
	if !hasMore {
		t.Error("期望还有更多数据")
	}

	// 场景3：正则过滤后分页
	filtered = FilterByName(pods, ".*-prod-.*", MatchRegex)
	if len(filtered) != 1 {
		t.Fatalf("期望过滤后 1 个 Pod，得到 %d 个", len(filtered))
	}

	page1, hasMore = Paginate(filtered, 5, 0)
	if len(page1) != 1 {
		t.Errorf("期望 1 个 Pod，得到 %d 个", len(page1))
	}
	if hasMore {
		t.Error("期望没有更多数据")
	}
}

// TestFilterByName_WithServices 测试用 Service 类型过滤
func TestFilterByName_WithServices(t *testing.T) {
	services := []*corev1.Service{
		{ObjectMeta: metav1.ObjectMeta{Name: "nginx-service"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "redis-service"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "mysql-service"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "web-frontend"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "web-backend"}},
	}

	result := FilterByName(services, "service")
	if len(result) != 3 {
		t.Errorf("期望 3 个 service，得到 %d 个", len(result))
	}

	result = FilterByName(services, "web", MatchPrefix)
	if len(result) != 2 {
		t.Errorf("期望 2 个 web 前缀的 service，得到 %d 个", len(result))
	}
}

// BenchmarkFilterByName 基准测试
func BenchmarkFilterByName(b *testing.B) {
	pods := createTestPods()

	b.Run("Contains", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FilterByName(pods, "web")
		}
	})

	b.Run("Prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FilterByName(pods, "nginx", MatchPrefix)
		}
	})

	b.Run("Suffix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FilterByName(pods, "-1", MatchSuffix)
		}
	})

	b.Run("Regex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FilterByName(pods, "^nginx-.*", MatchRegex)
		}
	})
}

// BenchmarkPaginate 分页基准测试
func BenchmarkPaginate(b *testing.B) {
	// 创建大列表
	items := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = i
	}

	b.Run("PageSize10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Paginate(items, 10, 0)
		}
	})

	b.Run("PageSize100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Paginate(items, 100, 0)
		}
	})
}
