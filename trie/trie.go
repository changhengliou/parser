package trie

type Node struct {
	Val      string
	Parent   *Node
	Children map[string]*Node
	term     bool
}

func New() *Node {
	return &Node{}
}

func (root *Node) Insert(val string) {
	node := root
	for _, c := range val {
		if childNode, exist := node.Children[string(c)]; exist {

		}
	}
}
