package common

import (
	"fmt"
	"testing"
)

var t *testing.T

func Assert(expect interface{}, actual interface{}) {
	if fmt.Sprintf("%v", expect) != fmt.Sprintf("%v", actual) {
		t.Errorf("Expect: %v, Actual: %v", expect, actual)
	}
}
