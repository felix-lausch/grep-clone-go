package main

import (
	"testing"
)

func TestMatchLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		pattern  string
		expected bool
	}{
		{
			name:     "Single rune match",
			line:     "hello",
			pattern:  "e",
			expected: true,
		},
		{
			name:     "Single rune no match",
			line:     "hello",
			pattern:  "z",
			expected: false,
		},
		{
			name:     "\\d matches digit",
			line:     "abc123",
			pattern:  "\\d",
			expected: true,
		},
		{
			name:     "\\d no match",
			line:     "abcdef",
			pattern:  "\\d",
			expected: false,
		},
		{
			name:     "\\w matches letter",
			line:     "hello!",
			pattern:  "\\w",
			expected: true,
		},
		{
			name:     "\\w matches underscore",
			line:     "___",
			pattern:  "\\w",
			expected: true,
		},
		{
			name:     "\\w no match",
			line:     "@@@@",
			pattern:  "\\w",
			expected: false,
		},
		{
			name:     "Character group match",
			line:     "abc",
			pattern:  "[ac]",
			expected: true,
		},
		{
			name:     "Character group match",
			line:     "a",
			pattern:  "[ac]",
			expected: true,
		},
		{
			name:     "Character group no match",
			line:     "xyz",
			pattern:  "[ab]",
			expected: false,
		},
		{
			name:     "Unsupported pattern",
			line:     "hello",
			pattern:  "(abc)",
			expected: false,
		},
		{
			name:     "Character negative group no match",
			line:     "apple",
			pattern:  "[^xyz]",
			expected: true,
		},
		{
			name:     "Character negative group match",
			line:     "apple",
			pattern:  "[^abc]",
			expected: true,
		},
		{
			name:     "Character negative group no match",
			line:     "abc",
			pattern:  "[^abc]",
			expected: false,
		},
		{
			name:     "Single Digit, multiple literals match",
			line:     "1 apple",
			pattern:  "\\d apple",
			expected: true,
		},
		{
			name:     "Single Digit, multiple literals no match",
			line:     "1 orange",
			pattern:  "\\d apple",
			expected: false,
		},
		{
			name:     "Multiple Digit, multiple literals match",
			line:     "100 apple",
			pattern:  "\\d\\d\\d apple",
			expected: true,
		},
		{
			name:     "Multiple Digit, multiple literals no match",
			line:     "1 apple",
			pattern:  "\\d\\d\\d apple",
			expected: false,
		},
		{
			name:     "1 Digit, multiple alphanumerics, 1 literal match",
			line:     "3 cats",
			pattern:  "\\d \\w\\w\\ws",
			expected: true,
		},
		{
			name:     "1 Digit, multiple alphanumerics, 1 literal match",
			line:     "what if i put it in the middle 3 dogs of my sentene",
			pattern:  "\\d \\w\\w\\ws",
			expected: true,
		},
		{
			name:     "1 Digit, multiple alphanumerics, 1 literal no match",
			line:     "1 dog",
			pattern:  "\\d \\w\\w\\ws",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchLine([]byte(tt.line), tt.pattern)

			if err != nil && tt.pattern[0] != '(' {
				t.Errorf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("matchLine(%q, %q) = %v; want %v", tt.line, tt.pattern, got, tt.expected)
			}
		})
	}
}
