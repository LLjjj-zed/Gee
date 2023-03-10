package gee

import "strings"

//使用 Trie 树实现动态路由(dynamic route)解析

// node Trie树 struct
// 与普通的树不同，为了实现动态路由匹配，加上了isWild这个参数。即当我们匹配 /p/go/doc/这个路由时，
// 第一层节点，p精准匹配到了p，第二层节点，go模糊匹配到:lang，那么将会把lang这个参数赋值为go，
// 继续下一层匹配。我们将匹配的逻辑，包装为一个辅助函数。
type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 时为true
}

// matchChild 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// matchChildren 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

//Trie 树需要支持节点的插入与查询。插入功能很简单，递归查找每一层的节点，
//如果没有匹配到当前part的节点，则新建一个，有一点需要注意，/p/:lang/doc只有在第三层节点，
//即doc节点，pattern才会设置为/p/:lang/doc。p和:lang节点的pattern属性皆为空。
//因此，当匹配结束时，我们可以使用n.pattern == ""来判断路由规则是否匹配成功。
//例如，/p/python虽能成功匹配到:lang，但:lang的pattern值为空，
//因此匹配失败。查询功能，同样也是递归查询每一层的节点，退出规则是，匹配到了*，匹配失败，或者匹配到了第len(parts)层节点。

// insert 这是一个使用递归方式实现的Trie树节点的插入方法，此方法用于将给定的模式字符串 pattern 插入到Trie树的节点 n 中。它接收三个参数：
//
//	pattern  是要插入的模式字符串。
//	parts    是 pattern 字符串按照 / 分割后的字符串切片。
//	height   是当前 Trie 树节点的高度，即该节点在 Trie 树中的深度。
//
// 如果字符串切片 parts 的长度等于路由树的高度 height，则表示已经到达了路由模式的末尾，可以将路由模式添加到当前节点中。
// 否则，从当前节点的子节点中查找与当前部分 part 匹配的节点，如果找到了匹配的子节点，则将其作为当前节点，递归调用 insert() 方法，
// 将剩余的部分添加到匹配的子节点中。如果没有找到匹配的子节点，则创建一个新的节点，将其作为当前节点的子节点，
// 将当前部分 part 添加到新节点中，然后递归调用 insert() 方法，将剩余的部分添加到新节点中。
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// search 用于在路由树中查找匹配给定路由部分的节点。
// 该方法接收两个参数：
//
//	parts   表示待匹配的路由部分，
//	height  表示当前节点在路由树中的高度。
//
// 在此方法中，首先判断字符串切片 parts 的长度是否等于路由树的高度 height，或者当前节点的 part 字段以 * 开头，如果是，
// 则表示已经匹配到了路由的末尾或通配符节点，此时判断当前节点是否包含路由模式，如果包含，则返回当前节点，否则返回 nil。
// 如果还没有匹配到路由的末尾，就从当前节点的子节点中查找与当前部分 part 匹配的节点，并递归递归调用 search() 方法，
// 匹配剩余的路由部分。如果找到了匹配的子节点，则将其作为当前节点，继续递归查找，直到匹配到路由的末尾或没有匹配的子节点。
// 如果没有匹配的子节点，则返回 nil。
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
