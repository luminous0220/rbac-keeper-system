package utils

// Build 将扁平的节点列表构建为树形结构，整体时间复杂度为 O(n)。
// 泛型 T 代表业务中的节点结构体。
//
// 参数说明:
//   - items: 扁平的节点切片
//   - getID: 提取节点主键的函数
//   - getParentID: 提取节点父键的函数
//   - attach: 将子节点指针切片挂回父节点的回调函数（由调用方实现，因为不同业务结构体的 Children 字段名不同）
//
// 注意事项:
//   - parentId 为 0 或父节点在列表中不存在时，该节点会被视为根节点（防止孤立节点丢失）。
//   - 调用方需负责在写入前校验循环引用；本函数仅按 parent 链单向挂载，不检测环。
func Build[T any](
	items []T,
	getID func(T) uint,
	getParentID func(T) uint,
	attach func(parent *T, children []*T),
) []T {
	// 如果输入为空，直接返回 nil
	if len(items) == 0 {
		return nil
	}

	// 创建一个完全独立的内存副本
	// 如果我们在 Build 函数内部直接对传入的 items 取地址（&items[i]），那么返回的树节点指针将直接指向调用方原始切片中的内存
	// 如果调用方在构建树之后，又去修改了原始的 items 切片，那么已经构建好的树结构也会跟着被篡改
	// 或者只要树节点还在被使用，整个原始的大切片就无法被垃圾回收器（GC）释放
	pool := make([]T, len(items))
	copy(pool, items)

	// 建立 "节点ID -> 节点指针" 的映射表，方便后续查找父节点
	byID := make(map[uint]*T, len(pool))
	for i := range pool {
		byID[getID(pool[i])] = &pool[i]
	}

	// 收集每个父节点对应的子节点指针切片，以及所有的根节点
	childrenOf := make(map[uint][]*T)
	var roots []*T

	for i := range pool {
		node := &pool[i]
		pid := getParentID(*node)

		// 如果 parentID 为 0，或者父节点在 byID 中不存在（孤立节点），则视为根节点
		if pid == 0 || byID[pid] == nil {
			roots = append(roots, node)
			continue
		}

		// 否则，将该节点追加到对应父节点的子节点列表中
		childrenOf[pid] = append(childrenOf[pid], node)
	}

	// 遍历子节点映射，调用 attach 回调将子节点切片挂载到真实的父节点上
	for pid, kids := range childrenOf {
		// byID[pid] 一定能找到，因为上面已经过滤了不存在的父节点
		attach(byID[pid], kids)
	}

	// 将根节点指针解引用，拷贝到新的切片中返回（避免暴露内部 pool 的指针）
	out := make([]T, len(roots))
	for i, r := range roots {
		out[i] = *r
	}
	return out
}

// 检测将 selfID 的父改为 newParentID 是否形成环。
// parentMap 为当前全部节点 id->parentID 映射表；沿 newParentID 向上追溯祖先
// 遇到 selfID 则成环（selfID 成为其自身后代的子节点）。
func WouldCreateCycle(parentMap map[uint]uint, selfID, newParentID uint) bool {
	if newParentID == 0 {
		return false // 挂到根不成环
	}
	if newParentID == selfID {
		return true // 自引用
	}
	// visited 防止 parentMap 数据本身已含环导致死循环
	// 假设你要把部门 100 挂到 200 下，但数据库里 200 和 300 已经形成了死循环（200的父是300，300的父是200）。
	// - 没有 visited：代码会在 200 和 300 之间无限往上找，导致服务崩溃。
	// - 有 visited：代码走过 200 和 300 后，发现再次遇到 200 时“脚印”已经存在，就会立刻意识到前面是个圈，从而安全终止循环
	visited := map[uint]struct{}{newParentID: {}}
	cur := newParentID

	for cur != 0 {
		p, ok := parentMap[cur]
		if !ok {
			return false // 父不在映射表，不成环
		}
		if p == selfID {
			return true
		}
		if _, seen := visited[p]; seen {
			return false // 数据已含环，避免死循环
		}
		visited[p] = struct{}{}
		cur = p
	}
	return false
}

// 收集指定根节点及其全部后代节点的 ID（BFS，包含根节点自身）。
// 用于"本部门及子部门"类的范围过滤（部门树选中某节点时，须连同其子部门一起参与筛选）。
// 注意事项：
//   - 调用方须传入全量节点列表（不应用数据权限过滤），否则子树会被权限裁剪、漏掉后代节点。
//   - rootID 为 0 或列表为空时返回 nil，调用方据此跳过过滤。
func CollectSubtreeIDs[T any](
	items []T,
	getID func(T) uint,
	getParentID func(T) uint,
	rootID uint,
) []uint {
	if rootID == 0 || len(items) == 0 {
		return nil
	}
	// 建立 parentID -> 子节点ID 的映射，自根向下 BFS，天然去重（每个节点只入队一次）
	childrenOf := make(map[uint][]uint, len(items))
	for i := range items {
		childrenOf[getParentID(items[i])] = append(childrenOf[getParentID(items[i])], getID(items[i]))
	}
	ids := make([]uint, 0, len(items))
	queue := []uint{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		ids = append(ids, cur)
		queue = append(queue, childrenOf[cur]...)
	}
	return ids
}

// 在树形结构中进行模糊搜索时，保留所有“命中条件的节点”，并自动把它们向上的“祖先节点（父级、爷爷级等）也一并保留”，从而保证返回给前端的数据依然是一棵结构完整的树，而不是散落的孤立节点
func FilterTreeWithAncestors[T any](
	items []T,
	getID func(T) uint,
	getParentID func(T) uint,
	match func(T) bool,
) []T {
	if len(items) == 0 {
		return items
	}
	byID := make(map[uint]*T, len(items))
	for i := range items {
		byID[getID(items[i])] = &items[i]
	}
	keep := make(map[uint]struct{})
	for i := range items {
		if !match(items[i]) {
			continue
		}
		// 从匹配节点向上收集祖先链
		cur := getID(items[i])
		for cur != 0 {
			keep[cur] = struct{}{}
			p, ok := byID[cur]
			if !ok {
				break
			}
			cur = getParentID(*p)
		}
	}
	out := make([]T, 0, len(keep))
	for i := range items {
		if _, ok := keep[getID(items[i])]; ok {
			out = append(out, items[i])
		}
	}
	return out
}
