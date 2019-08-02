package radixtree

import (
  "parser/common"
  "sort"
  "testing"
)

func TestLongestPrefix(t *testing.T) {
  common.Assert(3, longestPrefix("foobar", "foo"))
}

func TestLongestPrefix_2(t *testing.T) {
  common.Assert(0, longestPrefix("foobar", "bar"))

}

func TestLongestPrefix_3(t *testing.T) {
  common.Assert(0, longestPrefix("", "bar"))

}

func TestLongestPrefix_4(t *testing.T) {
  common.Assert(6, longestPrefix("foobarbar", "foobarfoo"))
}

func getNodes(arr []prefixStr) Nodes {
  result := make(Nodes, len(arr))
  for i, s := range arr {
    result[i] = &Node{data: s}
  }
  return result
}

func TestAddChild(t *testing.T) {
  tree := New()
  tree.root = &Node{
    children: getNodes([]prefixStr{"cutie", "facebook", "fade", "ruby"}),
  }
  sort.Sort(tree.root.children)

  tree.root.addChild(&Node{data: prefixStr("desk")})
  common.Assert("cutie", tree.root.children[0].data.str())
  common.Assert("desk", tree.root.children[1].data.str())
  common.Assert("facebook", tree.root.children[2].data.str())
  common.Assert("fade", tree.root.children[3].data.str())
  common.Assert("ruby", tree.root.children[4].data.str())
}

func TestAddChild_2(t *testing.T) {
  tree := New()
  tree.root = &Node{
    children: getNodes([]prefixStr{"cutie", "facebook", "fade", "ruby"}),
  }
  sort.Sort(tree.root.children)

  tree.root.addChild(&Node{data: prefixStr("facet")})
  tree.root.addChild(&Node{data: prefixStr("apple")})
  tree.root.addChild(&Node{data: prefixStr("zebra")})

  common.Assert("apple", tree.root.children[0].data.str())
  common.Assert("cutie", tree.root.children[1].data.str())
  common.Assert("facebook", tree.root.children[2].data.str())
  common.Assert("facet", tree.root.children[3].data.str())
  common.Assert("fade", tree.root.children[4].data.str())
  common.Assert("ruby", tree.root.children[5].data.str())
  common.Assert("zebra", tree.root.children[6].data.str())
}

func TestInsert(t *testing.T) {

}