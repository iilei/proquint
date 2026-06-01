package cli

import (
	"strings"
	"testing"
)

func TestFormatProquint(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "decimal 42",
			input:    "42",
			expected: "babop",
		},
		{
			name:     "hex max 256-bit",
			input:    "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			expected: strings.Repeat("zuzuz-", 15) + "zuzuz",
		},
		{
			name:     "decimal max 256-bit",
			input:    "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			expected: strings.Repeat("zuzuz-", 15) + "zuzuz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, ok := parseBigInt(tc.input)
			if !ok {
				t.Fatal("failed to parse test input")
			}
			if got := formatProquint(val, 0); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestFormatProquintPadGroups(t *testing.T) {
	val, ok := parseBigInt("42")
	if !ok {
		t.Fatal("failed to parse test input")
	}

	if got := formatProquint(val, 2); got != "babab-babop" {
		t.Fatalf("expected %q, got %q", "babab-babop", got)
	}
}

func TestDecodeProquint(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "babop to decimal 42",
			input:    "babop",
			expected: "42",
		},
		{
			name:     "babab-babop to decimal 42",
			input:    "babab-babop",
			expected: "42",
		},
		{
			name:     "zuzuz groups to max decimal",
			input:    strings.Repeat("zuzuz-", 15) + "zuzuz",
			expected: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := decodeProquint(tc.input)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if got := val.Text(10); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
