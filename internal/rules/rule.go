// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Rule interface {
	Matches(command commands.Command) bool
	At(command commands.Command) int
	Span(command commands.Command) int
	Reason(command commands.Command) string
	String() string
}

type described struct {
	reason      string
	description string
}

func (d described) Reason(commands.Command) string {
	return d.reason
}

func (d described) describes(written string) string {
	if d.description == "" {
		return written
	}
	return d.description
}
