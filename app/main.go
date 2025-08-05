package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	matched, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !matched {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {
	var ok bool

	if utf8.RuneCountInString(pattern) == 1 {
		ok = bytes.ContainsAny(line, pattern)
	} else if pattern == "\\d" {
		ok = bytes.ContainsFunc(line, func(r rune) bool {
			return unicode.IsDigit(r)
		})
	} else if pattern == "\\w" {
		ok = bytes.ContainsFunc(line, func(r rune) bool {
			return unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_'
		})
	} else if isCharacterGroup(pattern) {
		ok = checkCharacterGroup(line, pattern[1:len(pattern)-1])
	} else {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	return ok, nil
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")
}

func isCharacterGroup(pattern string) bool {
	return pattern[0] == '[' && pattern[len(pattern)-1] == ']'
}

func checkCharacterGroup(line []byte, group string) bool {
	positive := true
	if len(group) == 0 {
		return false
	} else if group[0] == '^' {
		positive = false
	}

	if !positive {
		return bytes.ContainsFunc(line, func(r rune) bool {
			return !strings.ContainsRune(group[1:], r)
		})
	}

	return bytes.ContainsAny(line, group)
}
