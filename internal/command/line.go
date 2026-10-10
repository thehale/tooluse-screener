// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import "strings"

const Substitution = "\x00"

func Blind(text string) []Command {
	return commands(asParts(split(withLinesJoined(text))), nil)
}

type part struct {
	text string
}

func asParts(texts []string) []part {
	parts := make([]part, len(texts))
	for index, text := range texts {
		parts[index] = part{text: text}
	}
	return parts
}

func commands(parts []part, startingDestinations []string) []Command {
	var found []Command
	destinations := startingDestinations
	for _, one := range parts {
		if text := withSpacesCollapsed(one.text); text != "" {
			found = append(found, Command{text, destinations})
			destinations = movesAfter(text, destinations)
		}
	}
	return found
}

func withLinesJoined(text string) string {
	return strings.ReplaceAll(text, "\\\n", "")
}

func split(text string) []string {
	outerText, substitutions := withoutSubstitutions(text)
	parts := splitAtSeparators(outerText)
	for _, substitution := range substitutions {
		parts = append(parts, split(substitution)...)
	}
	return parts
}

func withoutSubstitutions(text string) (outerText string, substitutions []string) {
	var runs strings.Builder
	for index := 0; index < len(text); {
		opener, isOpening := substitutionOpeningAt(text, index)
		if !isOpening {
			runs.WriteByte(text[index])
			index++
		} else {
			contents, next := readSubstitution(text, index, opener)
			substitutions = append(substitutions, contents)
			runs.WriteString(Substitution)
			index = next
		}
	}
	return runs.String(), substitutions
}

func substitutionOpeningAt(text string, index int) (opener string, isOpening bool) {
	rest := text[index:]
	if strings.HasPrefix(rest, arithmeticOpener) {
		return "", false
	}
	for _, candidate := range substitutionOpeners {
		if strings.HasPrefix(rest, candidate) {
			return candidate, true
		}
	}
	return "", false
}

const arithmeticOpener = "$(("

var substitutionOpeners = []string{"$(", "`", "<(", ">("}

func readSubstitution(text string, start int, opener string) (contents string, next int) {
	if opener == "`" {
		return readToClosingBacktick(text, start)
	} else {
		return readToClosingParen(text, start+len(opener)-1)
	}
}

func readToClosingBacktick(text string, start int) (contents string, next int) {
	begins := start + 1
	end := strings.Index(text[begins:], "`")
	if end < 0 {
		return restFrom(text, begins)
	} else {
		return text[begins : begins+end], begins + end + 1
	}
}

func readToClosingParen(text string, openingParen int) (contents string, next int) {
	begins, depth := openingParen+1, 0
	for index := openingParen; index < len(text); index++ {
		depth += nesting(text[index])
		if depth == 0 {
			return text[begins:index], index + 1
		}
	}
	return restFrom(text, begins)
}

func nesting(letter byte) int {
	switch letter {
	case '(':
		return 1
	case ')':
		return -1
	default:
		return 0
	}
}

func restFrom(text string, begins int) (contents string, next int) {
	return text[begins:], len(text)
}

func splitAtSeparators(text string) []string {
	var parts []string
	start := 0
	for index := 0; index < len(text); {
		if width := separatorAt(text, index); width == 0 {
			index++
		} else {
			parts = append(parts, text[start:index])
			index += width
			start = index
		}
	}
	return append(parts, text[start:])
}

func separatorAt(text string, index int) int {
	rest := text[index:]
	switch {
	case strings.HasPrefix(rest, "||"), strings.HasPrefix(rest, "&&"):
		return 2
	case strings.ContainsRune(";\n|", rune(text[index])):
		return 1
	case text[index] == '&' && isBackgrounding(text, index):
		return 1
	default:
		return 0
	}
}

func isBackgrounding(text string, index int) bool {
	return !strings.ContainsRune("<>&", letterAt(text, index-1)) &&
		!strings.ContainsRune("&>", letterAt(text, index+1))
}

func withSpacesCollapsed(command string) string {
	return strings.Join(strings.Fields(command), " ")
}

func letterAt(text string, index int) rune {
	if index < 0 || index >= len(text) {
		return 0
	} else {
		return rune(text[index])
	}
}
