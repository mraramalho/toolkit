package toolkit

import (
	"testing"
)

func TestRandomString(t *testing.T) {
	tools := new(Tools)
	str := tools.RandomString(10)

	if len(str) != 10 {
		t.Errorf("%s has %v letters", str, len(str))
	}

}
