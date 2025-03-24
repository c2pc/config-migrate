package replacer

import (
	"testing"
)

func TestRegisterTwice(t *testing.T) {
	Register("mock", func() string { return "" })

	var err interface{}
	func() {
		defer func() {
			err = recover()
		}()
		Register("mock", func() string { return "" })
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
		Register("mock", nil)
	}()

	if err == nil {
		t.Fatal("expected a panic when calling Register twice")
	}
}

func TestReplace(t *testing.T) {
	Register("__test_replacer__", func() string {
		return "-TEST_STRING-"
	})

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "Success",
			value:    "some_string__test_replacer__some_string",
			expected: "some_string-TEST_STRING-some_string",
		},
		{
			name:     "No replacer",
			value:    "some_string__some_string",
			expected: "some_string__some_string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Replace(tt.value)

			if result != tt.expected {
				t.Errorf("Expected\n %s, got\n %s", tt.expected, result)
			}
		})
	}
}
