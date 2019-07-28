package main

import "testing"

var t *testing.T

func assert(expect interface{}, actual interface{}) {
	if expect != actual {
		t.Errorf("Expect: %v, Actual: %v", expect, actual)
	}
}

func TestInitSymbols(t *testing.T) {
	p := New()

	p.initSymbols("T -> int Y | (E)")

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

	assert("int", symGrps[0][0].str())
	assert("Y", symGrps[0][1].str())
	assert("(E)", symGrps[1][0].str())
}