// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Rule interface {
	Matches(command commands.Command) bool
	At(command commands.Command) int
	Span(command commands.Command) int
	Note(command commands.Command) string
	String() string
}

type described struct {
	note        string
	description string
}

func (d described) Note(commands.Command) string {
	return d.note
}

func (d described) describes(written string) string {
	if d.description == "" {
		return written
	}
	return d.description
}
