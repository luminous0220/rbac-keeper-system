package utils

// 根据指定的字段对结构体切片进行去重，保留原始顺序。
// T: 切片元素的类型（例如 *model.Menu）
// K: 用于去重的键类型（例如 uint, string），必须支持比较操作
func UniqueBy[T any, K comparable](slice []T, getKey func(T) K) []T {
	if len(slice) == 0 {
		return nil
	}

	// 使用 map 记录已经出现过的键
	seen := make(map[K]struct{}, len(slice))
	// 预分配结果切片容量
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		key := getKey(item)
		// 如果该键没有被记录过
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
