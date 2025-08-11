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
	alternations, err := parseAlternations(pattern)
	if err != nil {
		return false, fmt.Errorf("error: parsing input pattern: %v", err)
	}

	for _, expressions := range alternations {
		matched, _ := match([]rune(string(line)), expressions)
		if matched {
			return true, nil
		}
	}

	return false, nil
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
	lineIndex := 0
	exprIndex := 0

	for exprIndex < len(expressions) {
		expr := expressions[exprIndex]

		if expr.Type == EndAnchor && exprIndex == len(expressions)-1 {
			return lineIndex == len(line)
		}

		if lineIndex >= len(line) {
			return hasMandatoryQuantities(expressions[exprIndex:])
		}

		switch expr.Quantity {
		case One: // Match exactly one rune
			if !matchExpression(line[lineIndex], expr) {
				return false
			}
			lineIndex++
			exprIndex++

		case OneOrMore: // Must match at least one
			if !matchExpression(line[lineIndex], expr) {
				return false
			}
			// Match as many as possible
			start := lineIndex
			for lineIndex < len(line) && matchExpression(line[lineIndex], expr) {
				lineIndex++
			}
			// Try all possible paths
			for split := lineIndex; split > start; split-- {
				if matchHere(line[split:], expressions[exprIndex+1:]) {
					return true
				}
			}
			return false

		case ZeroOrOne: // Match zero or one
			if matchExpression(line[lineIndex], expr) {
				lineIndex++
			}
			exprIndex++
		}
	}

	return true
}

func matchExpression(char rune, ex RegEx) bool {
	switch ex.Type {
	case Wildcard:
		return true
	case Literal:
		return char == ex.Char
	case Digit:
		return unicode.IsDigit(char)
	case AlphaNumeric:
		return unicode.IsDigit(char) || unicode.IsLetter(char) || char == '_'
	case Group:
		return checkCharacterGroup(char, ex.Group)
	default:
		return false
	}
}

func checkCharacterGroup(char rune, group string) bool {
	negative := group[0] == '^'

	if negative {
		return !strings.ContainsRune(group[1:], char)
	}

	return strings.ContainsRune(group, char)
}

func hasMandatoryQuantities(expressions []RegEx) bool {
	for _, ex := range expressions {
		if ex.Quantity != ZeroOrOne {
			return false
		}
	}

	return true
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
	Wildcard
)

const (
	One QuantityType = iota
	OneOrMore
	ZeroOrOne
)

type RegEx struct {
	Type     RegExType
	Group    string
	Char     rune
	Quantity QuantityType
}

func resolverInnerBrace(input []string) (bool, []string) {
	openingBraces := []int{}
	patterns := make(map[string]struct{})

	for _, pattern := range input {
		for i, r := range pattern {

			if r == '(' {
				openingBraces = append(openingBraces, i)
			}

			if r == ')' {
				openingBraceIdx := openingBraces[len(openingBraces)-1]
				closingBraceIdx := i

				//evaluate alternations and add options to patterns
				splitPattern := strings.SplitSeq(pattern[openingBraceIdx+1:closingBraceIdx], "|")

				for s := range splitPattern {
					pt := pattern[:openingBraceIdx] + s + pattern[closingBraceIdx+1:]
					patterns[pt] = struct{}{}
				}

				break
			}
		}
	}

	if len(patterns) > 0 {
		return true, keys(patterns)
	}

	return false, input
}

func keys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))

	for key := range m {
		result = append(result, key)
	}

	return result
}

func parseAlternations(input string) ([][]RegEx, error) {
	patterns := []string{input}

	for {
		hasMore, nextPatterns := resolverInnerBrace(patterns)
		patterns = nextPatterns

		if !hasMore {
			break
		}
	}

	result := [][]RegEx{}

	if len(patterns) == 0 {
		patterns = append(patterns, input)
	}

	for _, pt := range patterns {
		expressions, err := ParseExpressions(pt)
		if err != nil {
			return nil, fmt.Errorf("error: parsing input pattern: %v", err)
		}

		result = append(result, expressions)
	}

	return result, nil
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
		} else if s == '?' {
			result[len(result)-1].Quantity = ZeroOrOne
			continue
		} else if s == '.' {
			result = append(result, RegEx{Wildcard, "", '0', One})
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
