// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Words string

func (w Words) Find(command commands.Command) (Match, bool) {
	text := command.Text
	for from := 0; w != "" && from+len(w) <= len(text); {
		index := strings.Index(text[from:], string(w))
		switch {
		case index < 0:
			return Match{}, false
		case w.isStandaloneAt(text, from+index):
			return Match{from + index, len(w)}, true
		default:
			from += index + 1
		}
	}
	return Match{}, false
}

func (w Words) isStandaloneAt(command string, at int) bool {
	return w.isWordStartAt(command, at) && w.isWordEndAt(command, at)
}

func (w Words) isWordStartAt(command string, at int) bool {
	return at == 0 || !isWordLetter(command[at-1]) || !isWordLetter(w[0])
}

func (w Words) isWordEndAt(command string, at int) bool {
	after := at + len(w)
	return after == len(command) ||
		!isWordLetter(command[after]) ||
		!isWordLetter(w[len(w)-1])
}

func isWordLetter(letter byte) bool {
	return letter == '_' || letter == '-' ||
		('a' <= letter && letter <= 'z') ||
		('A' <= letter && letter <= 'Z') ||
		('0' <= letter && letter <= '9')
}

func (w Words) String() string {
	return string(w)
}
