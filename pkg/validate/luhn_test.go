package validate

import "testing"

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		number string
		valid  bool
	}{
		{"79927398713", true},      // valid
		{"12345678903", true},      // valid
		{"79927398710", false},     // invalid
		{"1234567890", false},      // invalid
		{"abcdef", false},          // not digits
		{"", false},                // empty
		{"0", false},               // too short
		{"4242424242424242", true}, // valid (card)
	}
	for _, tc := range tests {
		t.Run(tc.number, func(t *testing.T) {
			if IsValidLuhn(tc.number) != tc.valid {
				t.Errorf("IsValidLuhn(%q) = %v, want %v", tc.number, IsValidLuhn(tc.number), tc.valid)
			}
		})
	}
}
