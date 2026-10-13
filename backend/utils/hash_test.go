package utils

import (
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard url",
			input:    "https://example.com",
			expected: "a1b2c3d",
		},
		{
			name:     "different url",
			input:    "https://google.com",
			expected: "e4f5g6h",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "i7j8k9l",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash(tt.input)
			if len(result) != 7 {
				t.Errorf("Hash() length = %d, want 7", len(result))
			}
		})
	}
}

func TestHashDeterministic(t *testing.T) {
	input := "https://example.com/test"
	result1 := Hash(input)
	result2 := Hash(input)
	if result1 != result2 {
		t.Errorf("Hash() not deterministic: %s != %s", result1, result2)
	}
}

func TestHashDifferentInputs(t *testing.T) {
	input1 := "https://example.com/a"
	input2 := "https://example.com/b"
	result1 := Hash(input1)
	result2 := Hash(input2)
	if result1 == result2 {
		t.Errorf("Hash() collision for different inputs: %s == %s", result1, result2)
	}
}