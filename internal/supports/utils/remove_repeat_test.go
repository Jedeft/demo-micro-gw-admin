package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveUint32Duplication(t *testing.T) {
	tests := []struct {
		name  string
		input []uint32
		want  []uint32
	}{
		{name: "nil input", input: nil, want: nil},
		{name: "empty slice", input: []uint32{}, want: []uint32{}},
		{name: "no duplicates", input: []uint32{1, 2, 3}, want: []uint32{1, 2, 3}},
		{name: "with duplicates", input: []uint32{1, 2, 1, 3, 2}, want: []uint32{1, 2, 3}},
		{name: "all same", input: []uint32{5, 5, 5}, want: []uint32{5}},
		{name: "single element", input: []uint32{42}, want: []uint32{42}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveUint32Duplication(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveStringDuplication(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{name: "nil input", input: nil, want: nil},
		{name: "empty slice", input: []string{}, want: []string{}},
		{name: "no duplicates", input: []string{"a", "b", "c"}, want: []string{"a", "b", "c"}},
		{name: "with duplicates", input: []string{"a", "b", "a", "c", "b"}, want: []string{"a", "b", "c"}},
		{name: "all same", input: []string{"x", "x", "x"}, want: []string{"x"}},
		{name: "single element", input: []string{"hello"}, want: []string{"hello"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveStringDuplication(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
