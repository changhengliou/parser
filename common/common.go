package common

import (
	"fmt"
	"testing"
)

var t *testing.T

func Assert(expect interface{}, actual interface{}) {
	switch expect.(type) {
	case []string:
		e := expect.([]string)
		a := actual.([]string)
		if len(e) != len(a) {
			t.Errorf("Expect: %v, Actual: %v", expect, actual)
		}
		for i, obj := range e {
			if obj != a[i] {
				t.Errorf("Expect: %v, Actual: %v", expect, actual)
			}
		}
	case Set:
		e := expect.(Set)
		a := actual.(Set)
		if len(e) != len(a) {
			t.Errorf("Expect: %v, Actual: %v", e.Keys(), a.Keys())
		}
		for v := range e {
			if !a.Exist(v) {
				t.Errorf("Expect: %v, Actual: %v", e.Keys(), a.Keys())
			}
		}
	default:
		if fmt.Sprintf("%v", expect) != fmt.Sprintf("%v", actual) {
			t.Errorf("Expect: %v, Actual: %v", expect, actual)
		}
	}
}

type Set map[string]interface{}

func (s Set) Add(str string) {
	s[str] = nil
}

func (s Set) Exist(str string) bool {
	if _, exists := s[str]; exists {
		return true
	}
	return false
}

func (s Set) Keys() []string {
	i := 0
	arr := make([]string, len(s))
	for k := range s {
		arr[i] = k
		i++
	}
	return arr
}

func (s Set) AddSet(set2 Set) {
	for _, v := range set2.Keys() {
		s.Add(v)
	}
}