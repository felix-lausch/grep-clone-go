package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

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
	os.Exit(0)
}

func matchLine(line []byte, pattern string) (bool, error) {
	expressions, err := ParseExpressions(pattern)
	if err != nil {
		return false, fmt.Errorf("error: parsing input pattern: %v", err)
	}

	return match([]rune(string(line)), expressions)
}

func match(line []rune, expressions []RegEx) (bool, error) {
	for i := range line {
		if matchHere(line[i:], expressions) {
			return true, nil
		}
	}

	return false, nil
}

func matchHere(remainingLine []rune, expressions []RegEx) bool {
	//TODO: watch this, it will be wrong once more complex expressions are added
	if len(remainingLine) < len(expressions) {
		return false
	}

	for i := range expressions {
		if !matchExpression(remainingLine[i], expressions[i]) {
			return false
		}
	}

	return true
}

func matchExpression(char rune, ex RegEx) bool {
	var ok bool

	switch ex.Type {
	case Literal:
		ok = char == ex.Char
	case Digit:
		ok = unicode.IsDigit(char)
	case AlphaNumeric:
		ok = unicode.IsDigit(char) || unicode.IsLetter(char) || char == '_'
	case Group:
		ok = checkCharacterGroup(char, ex.Group)
	default:
		//TODO: i could return an error here for unknown ex Type. not really need atm
		ok = false
	}

	return ok
}

// func matchLine(line []byte, pattern string) (bool, error) {
// 	var ok bool

// 	//2. go through expressions 1 at a time and attempt to match them

// 	//3. only return 1 when all expressions could be matched

// 	if utf8.RuneCountInString(pattern) == 1 {
// 		ok = bytes.ContainsAny(line, pattern)
// 	} else if pattern == "\\d" {
// 		ok = bytes.ContainsFunc(line, func(r rune) bool {
// 			return unicode.IsDigit(r)
// 		})
// 	} else if pattern == "\\w" {
// 		ok = bytes.ContainsFunc(line, func(r rune) bool {
// 			return unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_'
// 		})
// 	} else if isCharacterGroup(pattern) {
// 		ok = checkCharacterGroup(line, pattern[1:len(pattern)-1])
// 	} else {
// 		return false, fmt.Errorf("unsupported pattern: %q", pattern)
// 	}

// 	return ok, nil
// 	// You can use print statements as follows for debugging, they'll be visible when running tests.
// 	// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")
// }

// func isCharacterGroup(pattern string) bool {
// 	return pattern[0] == '[' && pattern[len(pattern)-1] == ']'
// }

func checkCharacterGroup(char rune, group string) bool {
	negative := group[0] == '^'

	if negative {
		return !strings.ContainsRune(group[1:], char)
	}

	return strings.ContainsRune(group, char)
}

// func checkCharacterGroup(line []byte, group string) bool {
// 	positive := true
// 	if len(group) == 0 {
// 		return false
// 	} else if group[0] == '^' {
// 		positive = false
// 	}

// 	if !positive {
// 		return bytes.ContainsFunc(line, func(r rune) bool {
// 			return !strings.ContainsRune(group[1:], r)
// 		})
// 	}

// 	return bytes.ContainsAny(line, group)
// }

type RegExType int

const (
	Literal RegExType = iota
	Digit
	AlphaNumeric
	Group
)

type RegEx struct {
	Type  RegExType
	Group string
	Char  rune
}

func ParseExpressions(pattern string) ([]RegEx, error) {
	result := []RegEx{}
	escape := false
	activeGroup := false
	currentGroup := []rune{}

	for _, s := range pattern {
		if s == '\\' {
			escape = true
			continue
		} else if s == '[' {
			activeGroup = true
			continue
		} else if s == ']' {
			if len(currentGroup) == 0 {
				return nil, errors.New("invalid group snytax, provided empty group")
			}

			result = append(result, RegEx{Group, string(currentGroup), '0'})
			activeGroup = false
			currentGroup = []rune{}
			continue
		}

		if escape {
			switch s {
			case 'd':
				result = append(result, RegEx{Digit, "", '0'})
			case 'w':
				result = append(result, RegEx{AlphaNumeric, "", '0'})
			default:
				return nil, fmt.Errorf("escaped unknown character %v", s)
			}

			escape = false
			continue
		} else if activeGroup {
			if len(currentGroup) > 0 && s == '^' {
				return nil, errors.New("invalid group snytax, only first rune of group can be ^")
			}

			currentGroup = append(currentGroup, s)
			continue
		}

		result = append(result, RegEx{Literal, "", s})
	}

	return result, nil
}
