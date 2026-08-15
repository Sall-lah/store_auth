package sanitizer

import (
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mixed case with leading and trailing spaces",
			input:    "  User.Test@Example.COM  ",
			expected: "user.test@example.com",
		},
		{
			name:     "tabs and newlines around email",
			input:    "\t\n  ADMIN@SECURE.IO \r\n",
			expected: "admin@secure.io",
		},
		{
			name:     "already normalized email",
			input:    "clean@example.com",
			expected: "clean@example.com",
		},
		{
			name:     "empty string",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeEmail(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeEmail(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips script tags",
			input:    "<script>alert('xss')</script>John Doe",
			expected: "John Doe",
		},
		{
			name:     "strips nested or multiple html tags",
			input:    "<div><b>Jane</b> <i>Doe</i></div>",
			expected: "Jane Doe",
		},
		{
			name:     "removes non-printable control characters",
			input:    "Alice\x00\x07 Smith\x1F",
			expected: "Alice Smith",
		},
		{
			name:     "preserves international unicode characters and punctuation",
			input:    "  José O'Connor-Müller  ",
			expected: "José O'Connor-Müller",
		},
		{
			name:     "whitespace only",
			input:    "    ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trims whitespace and newline artifacts",
			input:    "  123456\n ",
			expected: "123456",
		},
		{
			name:     "strips embedded control characters",
			input:    "\x00987654\x1F",
			expected: "987654",
		},
		{
			name:     "clean otp code",
			input:    "654321",
			expected: "654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeCode(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeCode(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
