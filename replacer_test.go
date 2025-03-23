package migrate

import (
	"testing"
)

func TestRegisterTwice(t *testing.T) {
	RegisterReplacer("mock", func() string { return "" })

	var err interface{}
	func() {
		defer func() {
			err = recover()
		}()
		RegisterReplacer("mock", func() string { return "" })
	}()

	if err == nil {
		t.Fatal("expected a panic when calling Register twice")
	}
}

func TestRegisterEmpty(t *testing.T) {
	var err interface{}
	func() {
		defer func() {
			err = recover()
		}()
		RegisterReplacer("mock", nil)
	}()

	if err == nil {
		t.Fatal("expected a panic when calling Register twice")
	}
}
