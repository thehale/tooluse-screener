// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "strings"

type Words struct {
	described
	text string
}

func NewWords(text, note, description string) Rule {
	return Words{described{note, description}, text}
}

func (w Words) Matches(command string) bool {
	return w.At(command) >= 0
}

func (w Words) At(command string) int {
	for from := 0; w.text != "" && from+len(w.text) <= len(command); {
		found := strings.Index(command[from:], w.text)
		switch {
		case found < 0:
			return -1
		case w.standsAlone(command, from+found):
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

func (w Words) Span(command string) int {
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
