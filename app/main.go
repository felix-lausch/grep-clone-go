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

	if len(expressions) == 0 {
		return false, errors.New("can't match empty expression")
	}

	if expressions[0].Type == StartAnchor {
		return matchHere(line, expressions[1:]), nil
	}

	for i := range line {
		if matchHere(line[i:], expressions) {
			return true, nil
		}
	}

	return false, nil
}

func matchHere(line []rune, expressions []RegEx) bool {
	remainingLine := line

	for i := range expressions {

		if expressions[i].Type == EndAnchor && i == len(expressions)-1 {
			return len(line) == i
		}

		if i >= len(line) {
			return false
		}

		var matched bool

		matched, remainingLine = matchExpression(remainingLine, expressions[i])
		if !matched {
			return false
		}
	}

	return true
}

// func matchExpression(char rune, ex RegEx) bool {
// 	switch ex.Type {
// 	case Literal:
// 		return char == ex.Char
// 	case Digit:
// 		return unicode.IsDigit(char)
// 	case AlphaNumeric:
// 		return unicode.IsDigit(char) || unicode.IsLetter(char) || char == '_'
// 	case Group:
// 		return checkCharacterGroup(char, ex.Group)
// 	default:
// 		return false
// 	}
// }

func matchExpression(line []rune, ex RegEx) (bool, []rune) {
	if ex.Quantity == OneOrMore {
		foundOne := false
		remainingLine := line

		for {
			ex.Quantity = One
			matched := false
			matched, remainingLine = matchExpression(remainingLine, ex)

			if !matched {
				if foundOne {
					return true, remainingLine
				} else {
					return false, remainingLine
				}
			}

			if foundOne {
				return true, remainingLine
			}

			foundOne = true
		}
	}

	char := line[0]

	switch ex.Type {
	case Literal:
		res := char == ex.Char

		if res {
			return true, line[1:]
		} else {
			return false, line
		}

		// return char == ex.Char, line[1:]
	case Digit:
		return unicode.IsDigit(char), line[1:]
	case AlphaNumeric:
		return unicode.IsDigit(char) || unicode.IsLetter(char) || char == '_', line[1:]
	case Group:
		return checkCharacterGroup(char, ex.Group), line[1:]
	default:
		return false, line
	}
}

func checkCharacterGroup(char rune, group string) bool {
	negative := group[0] == '^'

	if negative {
		return !strings.ContainsRune(group[1:], char)
	}

	return strings.ContainsRune(group, char)
}

type RegExType int
type QuantityType int

const (
	Literal RegExType = iota
	Digit
	AlphaNumeric
	Group
	StartAnchor
	EndAnchor
)

const (
	One QuantityType = iota
	OneOrMore
)

type RegEx struct {
	Type     RegExType
	Group    string
	Char     rune
	Quantity QuantityType
}

func ParseExpressions(pattern string) ([]RegEx, error) {
	result := []RegEx{}
	escape := false
	groupActive := false
	currentGroup := []rune{}

	for _, s := range pattern {
		if s == '\\' {
			escape = true
			continue
		} else if s == '[' {
			groupActive = true
			continue
		} else if s == ']' {
			if len(currentGroup) == 0 {
				return nil, errors.New("invalid group snytax, provided empty group")
			}

			result = append(result, RegEx{Group, string(currentGroup), '0', One})
			groupActive = false
			currentGroup = []rune{}
			continue
		} else if s == '^' && !groupActive {
			result = append(result, RegEx{StartAnchor, "", '0', One})
			continue
		} else if s == '$' {
			result = append(result, RegEx{EndAnchor, "", '0', One})
			continue
		} else if s == '+' {
			result[len(result)-1].Quantity = OneOrMore
			continue
		}

		if escape {
			switch s {
			case 'd':
				result = append(result, RegEx{Digit, "", '0', One})
			case 'w':
				result = append(result, RegEx{AlphaNumeric, "", '0', One})
			default:
				return nil, fmt.Errorf("escaped unknown character %v", s)
			}

			escape = false
			continue
		} else if groupActive {
			if len(currentGroup) > 0 && s == '^' {
				return nil, errors.New("invalid group snytax, only first rune of group can be ^")
			}

			currentGroup = append(currentGroup, s)
			continue
		}

		result = append(result, RegEx{Literal, "", s, One})
	}

	return result, nil
}
