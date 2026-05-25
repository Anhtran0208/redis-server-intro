package data_structure

type Item struct {
	Score  float64
	Member string
}

func (item *Item) Compare(other *Item) int {
	if item.Score < other.Score {
		return -1
	}
	if item.Score > other.Score {
		return 1
	}

	if item.Member < other.Member {
		return -1
	}

	if item.Member > other.Member {
		return 1
	}
	return 0
}

type Node struct {
	Items    []*Item
	Children []*Node
	IsLeaf   bool
	Next     *Node
	Parent   *Node
}

func (node *Node) InsertItemToNode(item *Item) {
	i := 0
	for i < len(node.Items) && node.Items[i].Compare(item) < 0 {
		i++
	}
	node.Items = append(node.Items[:i], append([]*Item{item}, node.Items[i:]...)...)
}

type BPlusTree struct {
	Root      *Node
	Degree    int
	memberMap map[string]*Item
}

func NewBplusTree(degree int) *BPlusTree {
	return &BPlusTree{
		Root:      &Node{IsLeaf: true},
		Degree:    degree,
		memberMap: make(map[string]*Item),
	}
}

func (tree *BPlusTree) Add(score float64, member string) int {
	if member == "" {
		return 0
	}

	// member already exists
	if oldItem, exists := tree.memberMap[member]; exists {
		if oldItem.Score == score {
			return 0
		}

		oldLeaf := tree.findLeaf(oldItem)
		tree.removeItemFromLeaf(oldLeaf, oldItem)

		newItem := &Item{
			Score:  score,
			Member: member,
		}

		newLeaf := tree.findLeaf(newItem)
		newLeaf.InsertItemToNode(newItem)

		tree.memberMap[member] = newItem

		if tree.isOverflow(newLeaf) {
			if newLeaf == tree.Root {
				tree.splitRootLeaf()
			} else {
				tree.SplitNonRootLeaf(newLeaf)
			}
		}

		return 0
	}

	// new member
	item := &Item{
		Score:  score,
		Member: member,
	}
	leaf := tree.findLeaf(item)
	leaf.InsertItemToNode(item)

	tree.memberMap[member] = item

	// overflow
	if tree.isOverflow(leaf) {
		if leaf == tree.Root {
			tree.splitRootLeaf()
		} else {
			tree.SplitNonRootLeaf(leaf)
		}
	}
	return 1
}

func splitLeaf(node *Node) *Node {
	mid := len(node.Items) / 2
	newLeaf := &Node{
		IsLeaf: true,
		Next:   node.Next,
		Parent: node.Parent,
	}
	// after mid
	newLeaf.Items = append(newLeaf.Items, node.Items[mid:]...)
	//before mid
	node.Items = node.Items[:mid]
	node.Next = newLeaf
	return newLeaf
}

// when root is splited, create new root
func (tree *BPlusTree) splitRootLeaf() {
	currRoot := tree.Root
	newLeaf := splitLeaf(currRoot)

	newRoot := &Node{
		IsLeaf: false,
	}

	// first key of right leaf will be promoted to root
	newRoot.Items = append(newRoot.Items, newLeaf.Items[0])
	newRoot.Children = append(newRoot.Children, currRoot, newLeaf)

	currRoot.Parent = newRoot
	newLeaf.Parent = newRoot
	tree.Root = newRoot
}

func (tree *BPlusTree) SplitNonRootLeaf(leaf *Node) {
	currParent := leaf.Parent
	if currParent == nil {
		tree.splitRootLeaf()
		return
	}

	// split leaf
	newLeaf := splitLeaf(leaf)
	newLeaf.Parent = currParent

	// promote first item of new leaf to parent
	promotedItem := newLeaf.Items[0]

	// insert promoted items to parent.items
	tree.InsertPromotedItemToParent(currParent, leaf, promotedItem, newLeaf)

	// overflow
	if tree.isOverflow(currParent) {
		if currParent == tree.Root {
			tree.splitRootInternal()
		} else {
			tree.SplitNonRootInternal(currParent)
		}
	}
}

func (tree *BPlusTree) splitRootInternal() {
	oldRoot := tree.Root

	newRightNode, promotedItem := tree.splitInternalNode(oldRoot)

	newRoot := &Node{
		IsLeaf: false,
	}

	newRoot.Items = append(newRoot.Items, promotedItem)
	newRoot.Children = append(newRoot.Children, oldRoot, newRightNode)

	oldRoot.Parent = newRoot
	newRightNode.Parent = newRoot

	tree.Root = newRoot
}

func (tree *BPlusTree) SplitNonRootInternal(node *Node) {
	parent := node.Parent
	if parent == nil {
		tree.splitRootInternal()
		return
	}

	newRightNode, promotedItem := tree.splitInternalNode(node)

	tree.InsertPromotedItemToParent(parent, node, promotedItem, newRightNode)

	if tree.isOverflow(parent) {
		if parent == tree.Root {
			tree.splitRootInternal()
		} else {
			tree.SplitNonRootInternal(parent)
		}
	}
}

func (tree *BPlusTree) splitInternalNode(node *Node) (*Node, *Item) {
	mid := len(node.Items) / 2
	promotedItem := node.Items[mid]

	newRightNode := &Node{
		IsLeaf: false,
		Parent: node.Parent,
	}

	newRightNode.Items = append(newRightNode.Items, node.Items[mid+1:]...)
	newRightNode.Children = append(newRightNode.Children, node.Children[mid+1:]...)

	node.Items = node.Items[:mid]
	node.Children = node.Children[:mid+1]

	for _, child := range newRightNode.Children {
		child.Parent = newRightNode
	}

	return newRightNode, promotedItem
}

func (tree *BPlusTree) removeItemFromLeaf(leaf *Node, item *Item) bool {
	for i, currentItem := range leaf.Items {
		if currentItem.Compare(item) == 0 {
			leaf.Items = append(leaf.Items[:i], leaf.Items[i+1:]...)
			return true
		}
	}

	return false
}

func (tree *BPlusTree) InsertPromotedItemToParent(parent *Node, leftChild *Node, promotedItem *Item, rightChild *Node) {
	childIdx := tree.FindChildIndex(parent, leftChild)
	if childIdx == -1 {
		return
	}

	parent.Items = append(
		parent.Items[:childIdx],
		append([]*Item{promotedItem}, parent.Items[childIdx:]...)...,
	)

	parent.Children = append(
		parent.Children[:childIdx+1],
		append([]*Node{rightChild}, parent.Children[childIdx+1:]...)...,
	)

	rightChild.Parent = parent
}

func (tree *BPlusTree) FindChildIndex(parent *Node, child *Node) int {
	for i, currentChild := range parent.Children {
		if currentChild == child {
			return i
		}
	}
	return -1
}

func (tree *BPlusTree) findLeaf(item *Item) *Node {
	node := tree.Root
	for !node.IsLeaf {
		i := 0
		for i < len(node.Items) && item.Compare(node.Items[i]) >= 0 {
			i++
		}
		node = node.Children[i]
	}
	return node
}

func (tree *BPlusTree) isOverflow(node *Node) bool {
	return len(node.Items) > tree.Degree-1
}

func (tree *BPlusTree) Score(member string) (float64, bool) {
	item, exists := tree.memberMap[member]
	if !exists {
		return 0, false
	}

	return item.Score, true
}

func (tree *BPlusTree) GetRank(member string) int {
	if _, exists := tree.memberMap[member]; !exists {
		return -1
	}

	rank := 0

	node := tree.Root
	for !node.IsLeaf {
		node = node.Children[0]
	}

	for node != nil {
		for _, item := range node.Items {
			if item.Member == member {
				return rank
			}
			rank++
		}

		node = node.Next
	}

	return -1
}
