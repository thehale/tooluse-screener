// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import "strings"

const Unquotable = "\x00"

func All(text string) []string {
	var found []string
	for _, written := range split(joinedLines(text)) {
		if command := tidied(written); command != "" {
			found = append(found, command)
		}
	}
	return found
}

func joinedLines(text string) string {
	return strings.ReplaceAll(text, "\\\n", "")
}

func split(text string) []string {
	literal, substituted := withoutSubstitutions(text)
	found := atSeparators(literal)
	for _, substitution := range substituted {
		found = append(found, split(substitution)...)
	}
	return found
}

func withoutSubstitutions(text string) (literal string, substituted []string) {
	var runs strings.Builder
	for index := 0; index < len(text); {
		opener, opened := substitutionOpeningAt(text, index)
		if !opened {
			runs.WriteByte(text[index])
			index++
		} else {
			contents, resumed := readSubstitution(text, index, opener)
			substituted = append(substituted, contents)
			runs.WriteString(Unquotable)
			index = resumed
		}
	}
	return runs.String(), substituted
}

func substitutionOpeningAt(text string, index int) (opener string, opened bool) {
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

func readSubstitution(text string, start int, opener string) (contents string, resumed int) {
	if opener == "`" {
		return readToClosingBacktick(text, start)
	}
	return readToClosingParen(text, start+len(opener)-1)
}

func readToClosingBacktick(text string, start int) (contents string, resumed int) {
	begins := start + 1
	end := strings.Index(text[begins:], "`")
	if end < 0 {
		return unterminated(text, begins)
	}
	return text[begins : begins+end], begins + end + 1
}

func readToClosingParen(text string, openingParen int) (contents string, resumed int) {
	begins, depth := openingParen+1, 0
	for index := openingParen; index < len(text); index++ {
		depth += nesting(text[index])
		if depth == 0 {
			return text[begins:index], index + 1
		}
	}
	return unterminated(text, begins)
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

func unterminated(text string, begins int) (contents string, resumed int) {
	return text[begins:], len(text)
}

func atSeparators(text string) []string {
	var found []string
	start := 0
	for index := 0; index < len(text); {
		if width := separatorAt(text, index); width == 0 {
			index++
		} else {
			found = append(found, text[start:index])
			index += width
			start = index
		}
	}
	return append(found, text[start:])
}

func separatorAt(text string, index int) int {
	rest := text[index:]
	switch {
	case strings.HasPrefix(rest, "||"), strings.HasPrefix(rest, "&&"):
		return 2
	case strings.ContainsRune(";\n|", rune(text[index])):
		return 1
	case text[index] == '&' && backgrounds(text, index):
		return 1
	default:
		return 0
	}
}

func backgrounds(text string, index int) bool {
	return !strings.ContainsRune("<>&", letterAt(text, index-1)) &&
		!strings.ContainsRune("&>", letterAt(text, index+1))
}

func tidied(command string) string {
	return strings.Join(strings.Fields(command), " ")
}

func letterAt(text string, index int) rune {
	if index < 0 || index >= len(text) {
		return 0
	}
	return rune(text[index])
}
