package common

import "testing"

var t *testing.T

func Assert(expect interface{}, actual interface{}) {
  if expect != actual {
    t.Errorf("Expect: %v, Actual: %v", expect, actual)
  }
}