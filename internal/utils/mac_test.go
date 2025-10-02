package utils

import "testing"

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colons lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "dashes lowercase",
			input:    "aa-bb-cc-dd-ee-ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "mixed case colons",
			input:    "Aa:Bb:Cc:Dd:Ee:Ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "mixed case dashes",
			input:    "Aa-Bb-Cc-Dd-Ee-Ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "already uppercase colons",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeMAC(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeMAC(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidMAC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid colons lowercase",
			input: "aa:bb:cc:dd:ee:ff",
			valid: true,
		},
		{
			name:  "valid colons uppercase",
			input: "AA:BB:CC:DD:EE:FF",
			valid: true,
		},
		{
			name:  "valid dashes lowercase",
			input: "aa-bb-cc-dd-ee-ff",
			valid: true,
		},
		{
			name:  "valid dashes uppercase",
			input: "AA-BB-CC-DD-EE-FF",
			valid: true,
		},
		{
			name:  "mixed case valid",
			input: "Aa:Bb:Cc:Dd:Ee:Ff",
			valid: true,
		},
		{
			name:  "invalid too short",
			input: "aa:bb:cc:dd:ee",
			valid: false,
		},
		{
			name:  "invalid too long",
			input: "aa:bb:cc:dd:ee:ff:gg",
			valid: false,
		},
		{
			name:  "invalid characters",
			input: "gg:bb:cc:dd:ee:ff",
			valid: false,
		},
		{
			name:  "invalid format",
			input: "aabbccddeeff",
			valid: false,
		},
		{
			name:  "empty string",
			input: "",
			valid: false,
		},
		{
			name:  "invalid mixed separators",
			input: "aa:bb-cc:dd:ee:ff",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMAC(tt.input)
			if result != tt.valid {
				t.Errorf("IsValidMAC(%s) = %t, want %t", tt.input, result, tt.valid)
			}
		})
	}
}
