package radixtree

import (
  "sort"
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

    splitChild := &Node{
      data:     curr.data[commonPrefix:],
      children: curr.children,
      parent:   curr,
      term:     true,
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
