package main

import (
	"parser/common"
	"testing"
)

var t *testing.T



func TestInitSymbols(t *testing.T) {
	p := New()

	p.initSymbols("T -> int Y | (E)", nil)

	if len(p.symMap) != 1 || p.symMap["T"] == nil {
		t.Error("Expect to have 1 key T")
	}

	symGrps := p.symMap["T"]
	if len(symGrps) != 2 {
		t.Error("Expect to have 2 symbolGroups")
	}
	if len(symGrps[0]) != 2 {
		t.Error("Expect to have 2 symbols in group 0")
	}
	if len(symGrps[1]) != 1 {
		t.Error("Expect to have 1 symbols in group 1")
	}

	common.Assert("int", symGrps[0][0])
	common.Assert("Y", symGrps[0][1])
	common.Assert("(E)", symGrps[1][0])
}