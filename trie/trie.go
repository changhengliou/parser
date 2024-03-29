package main

import (
<<<<<<< HEAD
  "parser/radixtree"
  "sort"
  "strings"
)

// WalkFn is used when walking the tree. Takes a
// key and value, returning if iteration should
// be terminated.
type WalkFn func(s string, v interface{}) bool

// leafNode is used to represent a value
type leafNode struct {
  key string
  val interface{}
}

// edge is used to represent an edge node
type edge struct {
  label byte
  node  *node
}

type node struct {
  // leaf is used to store possible leaf
  leaf *leafNode

  // prefix is the common prefix we ignore
  prefix string

  // Edges should be stored in-order for iteration.
  // We avoid a fully materialized slice to save memory,
  // since in most cases we expect to be sparse
  edges edges
}

func (n *node) isLeaf() bool {
  return n.leaf != nil
}

func (n *node) addEdge(e edge) {
  n.edges = append(n.edges, e)
  n.edges.Sort()
}

func (n *node) updateEdge(label byte, node *node) {
  num := len(n.edges)
  idx := sort.Search(num, func(i int) bool {
    return n.edges[i].label >= label
  })
  if idx < num && n.edges[idx].label == label {
    n.edges[idx].node = node
    return
  }
  panic("replacing missing edge")
}

func (n *node) getEdge(label byte) *node {
  num := len(n.edges)
  idx := sort.Search(num, func(i int) bool {
    return n.edges[i].label >= label
  })
  if idx < num && n.edges[idx].label == label {
    return n.edges[idx].node
  }
  return nil
}

func (n *node) delEdge(label byte) {
  num := len(n.edges)
  idx := sort.Search(num, func(i int) bool {
    return n.edges[i].label >= label
  })
  if idx < num && n.edges[idx].label == label {
    copy(n.edges[idx:], n.edges[idx+1:])
    n.edges[len(n.edges)-1] = edge{}
    n.edges = n.edges[:len(n.edges)-1]
  }
}

type edges []edge

func (e edges) Len() int {
  return len(e)
}

func (e edges) Less(i, j int) bool {
  return e[i].label < e[j].label
}

func (e edges) Swap(i, j int) {
  e[i], e[j] = e[j], e[i]
}

func (e edges) Sort() {
  sort.Sort(e)
}

// Tree implements a radix tree. This can be treated as a
// Dictionary abstract data type. The main advantage over
// a standard hash map is prefix-based lookups and
// ordered iteration,
type Tree struct {
  root *node
  size int
}

// New returns an empty Tree
func New() *Tree {
  return NewFromMap(nil)
}

// NewFromMap returns a new tree containing the keys
// from an existing map
func NewFromMap(m map[string]interface{}) *Tree {
  t := &Tree{root: &node{}}
  for k, v := range m {
    t.Insert(k, v)
  }
  return t
}

// Len is used to return the number of elements in the tree
func (t *Tree) Len() int {
  return t.size
}

// longestPrefix finds the length of the shared prefix
// of two strings
func longestPrefix(k1, k2 string) int {
  max := len(k1)
  if l := len(k2); l < max {
    max = l
  }
  var i int
  for i = 0; i < max; i++ {
    if k1[i] != k2[i] {
      break
    }
  }
  return i
}

// Insert is used to add a newentry or update
// an existing entry. Returns if updated.
func (t *Tree) Insert(s string, v interface{}) (interface{}, bool) {
  var parent *node
  rootNode := t.root
  search := s
  for {
    // Handle key exhaustion
    if len(search) == 0 {
      if rootNode.isLeaf() {
        old := rootNode.leaf.val
        rootNode.leaf.val = v
        return old, true
      }

      rootNode.leaf = &leafNode{
        key: s,
        val: v,
      }
      t.size++
      return nil, false
    }

    // Look for the edge
    parent = rootNode
    rootNode = rootNode.getEdge(search[0])

    // No edge, create one
    if rootNode == nil {
      e := edge{
        label: search[0],
        node: &node{
          leaf: &leafNode{
            key: s,
            val: v,
          },
          prefix: search,
        },
      }
      parent.addEdge(e)
      t.size++
      return nil, false
    }

    // Determine longest prefix of the search key on match
    commonPrefix := longestPrefix(search, rootNode.prefix)
    if commonPrefix == len(rootNode.prefix) {
      search = search[commonPrefix:]
      continue
    }

    // Split the node
    t.size++
    child := &node{
      prefix: search[:commonPrefix],
    }
    parent.updateEdge(search[0], child)

    // Restore the existing node
    child.addEdge(edge{
      label: rootNode.prefix[commonPrefix],
      node:  rootNode,
    })
    rootNode.prefix = rootNode.prefix[commonPrefix:]

    // Create a new leaf node
    leaf := &leafNode{
      key: s,
      val: v,
    }

    // If the new key is a subset, add to to this node
    search = search[commonPrefix:]
    if len(search) == 0 {
      child.leaf = leaf
      return nil, false
    }

    // Create a new edge for the node
    child.addEdge(edge{
      label: search[0],
      node: &node{
        leaf:   leaf,
        prefix: search,
      },
    })
    return nil, false
  }
}

// Delete is used to delete a key, returning the previous
// value and if it was deleted
func (t *Tree) Delete(s string) (interface{}, bool) {
  var (
    parent *node
    curr   = t.root
    label  byte
    search = s
  )

  for {
    // Check for key exhaustion
    if len(search) == 0 {
      if !curr.isLeaf() {
        return nil, false
      }
      break
    }

    // Look for an edge
    label = search[0]
    parent = curr
    curr = curr.getEdge(label)
    if curr == nil {
      return nil, false
    }

    // Consume the search prefix
    if strings.HasPrefix(search, curr.prefix) {
      search = search[len(curr.prefix):]
    } else {
      return nil, false
    }
  }
  // Delete the leaf
  leaf := curr.leaf
  curr.leaf = nil
  t.size--

  // Check if we should delete this node from the parent
  if parent != nil && len(curr.edges) == 0 {
    parent.delEdge(label)
  }

  // Check if we should merge this node
  if curr != t.root && len(curr.edges) == 1 {
    curr.mergeChild()
  }

  // Check if we should merge the parent's other child
  if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
    parent.mergeChild()
  }

  return leaf.val, true
}

// DeletePrefix is used to delete the subtree under a prefix
// Returns how many nodes were deleted
// Use this to delete large subtrees efficiently
func (t *Tree) DeletePrefix(s string) int {
  return t.deletePrefix(nil, t.root, s)
}

// delete does a recursive deletion
func (t *Tree) deletePrefix(parent, n *node, prefix string) int {
  // Check for key exhaustion
  if len(prefix) == 0 {
    // Remove the leaf node
    subTreeSize := 0
    //recursively walk from all edges of the node to be deleted
    recursiveWalk(n, func(s string, v interface{}) bool {
      subTreeSize++
      return false
    })
    if n.isLeaf() {
      n.leaf = nil
    }
    n.edges = nil // deletes the entire subtree

    // Check if we should merge the parent's other child
    if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
      parent.mergeChild()
    }
    t.size -= subTreeSize
    return subTreeSize
  }

  // Look for an edge
  label := prefix[0]
  child := n.getEdge(label)
  if child == nil || (!strings.HasPrefix(child.prefix, prefix) && !strings.HasPrefix(prefix, child.prefix)) {
    return 0
  }

  // Consume the search prefix
  if len(child.prefix) > len(prefix) {
    prefix = prefix[len(prefix):]
  } else {
    prefix = prefix[len(child.prefix):]
  }
  return t.deletePrefix(n, child, prefix)
}

func (n *node) mergeChild() {
  e := n.edges[0]
  child := e.node
  n.prefix = n.prefix + child.prefix
  n.leaf = child.leaf
  n.edges = child.edges
}

// Get is used to lookup a specific key, returning
// the value and if it was found
func (t *Tree) Get(s string) (interface{}, bool) {
  n := t.root
  search := s
  for {
    // Check for key exhaution
    if len(search) == 0 {
      if n.isLeaf() {
        return n.leaf.val, true
      }
      break
    }

    // Look for an edge
    n = n.getEdge(search[0])
    if n == nil {
      break
    }

    // Consume the search prefix
    if strings.HasPrefix(search, n.prefix) {
      search = search[len(n.prefix):]
    } else {
      break
    }
  }
  return nil, false
}

// LongestPrefix is like Get, but instead of an
// exact match, it will return the longest prefix match.
func (t *Tree) LongestPrefix(s string) (string, interface{}, bool) {
  var last *leafNode
  n := t.root
  search := s
  for {
    // Look for a leaf node
    if n.isLeaf() {
      last = n.leaf
    }

    // Check for key exhaustion
    if len(search) == 0 {
      break
    }

    // Look for an edge
    n = n.getEdge(search[0])
    if n == nil {
      break
    }

    // Consume the search prefix
    if strings.HasPrefix(search, n.prefix) {
      search = search[len(n.prefix):]
    } else {
      break
    }
  }
  if last != nil {
    return last.key, last.val, true
  }
  return "", nil, false
}

// Minimum is used to return the minimum value in the tree
func (t *Tree) Minimum() (string, interface{}, bool) {
  n := t.root
  for {
    if n.isLeaf() {
      return n.leaf.key, n.leaf.val, true
    }
    if len(n.edges) > 0 {
      n = n.edges[0].node
    } else {
      break
    }
  }
  return "", nil, false
}

// Maximum is used to return the maximum value in the tree
func (t *Tree) Maximum() (string, interface{}, bool) {
  n := t.root
  for {
    if num := len(n.edges); num > 0 {
      n = n.edges[num-1].node
      continue
    }
    if n.isLeaf() {
      return n.leaf.key, n.leaf.val, true
    }
    break
  }
  return "", nil, false
}

// Walk is used to walk the tree
func (t *Tree) Walk(fn WalkFn) {
  recursiveWalk(t.root, fn)
}

// WalkPrefix is used to walk the tree under a prefix
func (t *Tree) WalkPrefix(prefix string, fn WalkFn) {
  n := t.root
  search := prefix
  for {
    // Check for key exhaution
    if len(search) == 0 {
      recursiveWalk(n, fn)
      return
    }

    // Look for an edge
    n = n.getEdge(search[0])
    if n == nil {
      break
    }

    // Consume the search prefix
    if strings.HasPrefix(search, n.prefix) {
      search = search[len(n.prefix):]

    } else if strings.HasPrefix(n.prefix, search) {
      // Child may be under our search prefix
      recursiveWalk(n, fn)
      return
    } else {
      break
    }
  }

}

// WalkPath is used to walk the tree, but only visiting nodes
// from the root down to a given leaf. Where WalkPrefix walks
// all the entries *under* the given prefix, this walks the
// entries *above* the given prefix.
func (t *Tree) WalkPath(path string, fn WalkFn) {
  n := t.root
  search := path
  for {
    // Visit the leaf values if any
    if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
      return
    }

    // Check for key exhaution
    if len(search) == 0 {
      return
    }

    // Look for an edge
    n = n.getEdge(search[0])
    if n == nil {
      return
    }

    // Consume the search prefix
    if strings.HasPrefix(search, n.prefix) {
      search = search[len(n.prefix):]
    } else {
      break
    }
  }
}

// recursiveWalk is used to do a pre-order walk of a node
// recursively. Returns true if the walk should be aborted
func recursiveWalk(n *node, fn WalkFn) bool {
  // Visit the leaf values if any
  if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
    return true
  }

  // Recurse on the children
  for _, e := range n.edges {
    if recursiveWalk(e.node, fn) {
      return true
    }
  }
  return false
=======
	"sort"
	"strings"
	"unicode/utf8"
)

type Nodes []*Node

func (n Nodes) Len() int {
	return len(n)
}

func (n Nodes) Less(i, j int) bool {
	return n[i].data.str() < n[j].data.str()
}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

type prefixStr string

func (p prefixStr) str() string {
	return string(p)
}

func (p prefixStr) firstChar() rune {
	r, _ := utf8.DecodeRuneInString(string(p))
	return r
}

type Tree struct {
	root *Node
}

type Node struct {
	data     prefixStr
	children Nodes
	parent   *Node
	term     bool
}

func (t *Node) findNode(label rune) *Node {
	childNum := len(t.children)
	index := sort.Search(childNum, func(i int) bool {
		return t.children[i].data.firstChar() >= label
	})
	if index < childNum && t.children[index].data.firstChar() == label {
		return t.children[index]
	}
	return nil
}

func (t *Node) addChild(n *Node) {
	childCount := len(t.children)
	label := n.data.str()
	index := sort.Search(childCount, func(i int) bool {
		return t.children[i].data.str() >= label
	})
	if index <= childCount {
		t.children = append(t.children[:index], append([]*Node{n}, t.children[index:]...)...)
	}
}

func (t *Node) removeChild(n *Node) {
	label := n.data.firstChar()
	childCount := len(t.children)
	index := sort.Search(childCount, func(i int) bool {
		return t.children[i].data.firstChar() >= label
	})
	if index < childCount && t.children[index].data.firstChar() == label {
		t.children = append(t.children[:index], t.children[index+1:]...)
	}
}

func New() *Tree {
	return &Tree{
		&Node{children: make([]*Node, 0)},
	}
}

// Insert:
// 1. no prefix matching -> direct insertion
// 2. full matching with existed word -> split existed word
// 3. partially matching with existed word -> branching
func (t *Tree) Insert(str string) {
	var (
		parent *Node
		curr   = t.root
		search = str
	)
	for {
		parent = curr
		curr = parent.findNode(prefixStr(search).firstChar())

		// no edge
		if curr == nil {
			newNode := &Node{
				data:     prefixStr(search),
				children: make([]*Node, 0),
				parent:   parent,
				term:     true,
			}
			parent.addChild(newNode)
			return
		}

		// insert: facebook, curr: face -> insert book under 'face'
		commonPrefix := longestPrefix(search, curr.data.str())
		// full matching with existed word, split here
		if commonPrefix == len(curr.data.str()) {
			search = search[commonPrefix:]
			continue
		}

		// fooloop foreach, splitChild = loop, true
		// for -> each
		//     -> loop
		// f -> or
		//   -> ace
		splitChildTerm := curr.term
		splitChild := &Node{
			data:     curr.data[commonPrefix:],
			children: curr.children,
			parent:   curr,
			term:     splitChildTerm,
		}
		curr.children = Nodes{splitChild}

		if splitSearch := search[commonPrefix:]; len(splitSearch) != 0 {
			curr.addChild(&Node{
				data:     prefixStr(splitSearch),
				children: make([]*Node, 0),
				parent:   curr,
				term:     true,
			})
			curr.term = false
		} else {
			curr.term = true
		}
		curr.data = curr.data[:commonPrefix]
		return
	}
}

func longestPrefix(a string, b string) int {
	i := 0
	for ; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return i
}

func (t *Tree) Remove(s string) bool {
	var (
		parent *Node
		curr   = t.root
		search = s
	)

	for {
		if len(search) == 0 {
			if !curr.term {
				return false
			}
			break
		}
		parent = curr
		curr = curr.findNode(rune(search[0]))
		if curr == nil {
			return false
		}

		if strings.HasPrefix(search, curr.data.str()) {
			search = search[len(curr.data):]
			continue
		}
		return false
	}
	curr.term = false
	if len(curr.children) == 0 {
		parent.removeChild(curr)
	}

	// merge nodes
	if len(curr.children) == 1 {
	  curr.term = curr.children[0].term
		curr.data += curr.children[0].data
    curr.children = curr.children[0].children
	}
	// check if we should merge the parent's other child
	return true
}

func (t *Tree) RemovePrefix(s string) {
	t.removePrefix(nil, t.root, s)
}

func (t *Tree) removePrefix(parent, child *Node, prefix string) {
	if len(prefix) == 0 {
		child.children = Nodes{}
	}

	curr := child.findNode(child.data.firstChar())
	if curr == nil {
		return
	}
	if len(curr.data) > len(prefix) {
		prefix = prefix[len(prefix):]
	} else {
		prefix = prefix[len(child.data):]
	}
	t.removePrefix(child, curr, prefix)
>>>>>>> 7c8c4869a275dcf582e07a496530169ed065f57f
}

// ToMap is used to walk the tree and convert it into a map
func (t *Tree) ToMap() map[string]interface{} {
  out := make(map[string]interface{}, t.size)
  t.Walk(func(k string, v interface{}) bool {
    out[k] = v
    return false
  })
  return out
}

func main() {
  t := New()
  t.Insert("forloop", 0)
  t.Insert("foreach", 1)
  t.Insert("face", 2)
  t.Insert("facebook", 3)
  t.Insert("facebook groups", 4)
  t.Insert("facebook group", 5)
  t.Delete("foe")

  tree := radixtree.New()
  tree.Insert("forloop")
  tree.Insert("foreach")
  tree.Insert("face")
  tree.Insert("facebook")
  tree.Insert("facebook groups")
  tree.Insert("facebook group")
}

// f -> or  -> loop
//          -> each
//   -> ace
//          \-> book
