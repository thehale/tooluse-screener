// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Words struct {
	wording
	text string
}

func NewWords(text, reason, name string) Rule {
	return Words{wording{reason, name}, text}
}

func (w Words) At(command commands.Command) int {
	text := command.Text
	for from := 0; w.text != "" && from+len(w.text) <= len(text); {
		index := strings.Index(text[from:], w.text)
		switch {
		case index < 0:
			return -1
		case w.isStandaloneAt(text, from+index):
			return from + index
		default:
			from += index + 1
		}
	}
	return -1
}

func (w Words) isStandaloneAt(command string, at int) bool {
	return w.isWordStartAt(command, at) && w.isWordEndAt(command, at)
}

func (w Words) isWordStartAt(command string, at int) bool {
	return at == 0 || !isWordLetter(command[at-1]) || !isWordLetter(w.text[0])
}

func (w Words) isWordEndAt(command string, at int) bool {
	after := at + len(w.text)
	return after == len(command) ||
		!isWordLetter(command[after]) ||
		!isWordLetter(w.text[len(w.text)-1])
}

func isWordLetter(letter byte) bool {
	return letter == '_' || letter == '-' ||
		('a' <= letter && letter <= 'z') ||
		('A' <= letter && letter <= 'Z') ||
		('0' <= letter && letter <= '9')
}

func (w Words) Span(command commands.Command) int {
	if w.At(command) >= 0 {
		return len(w.text)
	} else {
		return 0
	}
}

func (w Words) String() string {
	return w.nameOr(w.text)
}
