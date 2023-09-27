package node

import "testing"

func TestAdd(t *testing.T) {
	someNumber := 1
	someNumber *= 3
	if someNumber != 3 {
		t.Errorf("We failed")
	}
}
