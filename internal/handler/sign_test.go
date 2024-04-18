package handler

import (
	"fmt"
	"testing"
)

type customType struct{}

func (c customType) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal customType to JSON")
}

func TestSign(t *testing.T) {
	tt := []struct {
		name     string
		val      any
		expected string
	}{
		{"good case", "value", "f1c43f70384a385e8447450f39bff4b5ad125312c0d7d8a85569800d57f33d62"},
		{"bad case", customType{}, ""},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			got := sign(test.val, "key")
			if got != test.expected {
				t.Errorf("expected: %s, got: %s", test.expected, got)
			}
		})
	}
}
