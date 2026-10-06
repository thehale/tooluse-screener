// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Words struct {
	described
	text string
}

func NewWords(text, reason, description string) Rule {
	return Words{described{reason, description}, text}
}

func (w Words) Matches(command commands.Command) bool {
	return w.At(command) >= 0
}

func (w Words) At(command commands.Command) int {
	text := command.Text
	for from := 0; w.text != "" && from+len(w.text) <= len(text); {
		found := strings.Index(text[from:], w.text)
		switch {
		case found < 0:
			return -1
		case w.standsAlone(text, from+found):
			return from + found
		default:
			from += found + 1
		}
	}
	return -1
}

func (w Words) standsAlone(command string, at int) bool {
	return w.opensAWord(command, at) && w.closesAWord(command, at)
}

func (w Words) opensAWord(command string, at int) bool {
	return at == 0 || !joinsAWord(command[at-1]) || !joinsAWord(w.text[0])
}

func (w Words) closesAWord(command string, at int) bool {
	after := at + len(w.text)
	return after == len(command) ||
		!joinsAWord(command[after]) ||
		!joinsAWord(w.text[len(w.text)-1])
}

func joinsAWord(letter byte) bool {
	return letter == '_' || letter == '-' ||
		('a' <= letter && letter <= 'z') ||
		('A' <= letter && letter <= 'Z') ||
		('0' <= letter && letter <= '9')
}

func (w Words) Span(command commands.Command) int {
	switch {
	case w.Matches(command):
		return len(w.text)
	default:
		return 0
	}
}

func (w Words) String() string {
	return w.describes(w.text)
}
