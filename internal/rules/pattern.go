// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Pattern struct {
	described
	expression *regexp.Regexp
	written    string
}

func NewPattern(expression, note, description string) (Rule, error) {
	built, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	return Pattern{described{note, description}, built, expression}, nil
}

func (p Pattern) Matches(command commands.Command) bool {
	return p.At(command) >= 0
}

func (p Pattern) At(command commands.Command) int {
	return opened(p.expression.FindStringIndex(command.Text))
}

func opened(at []int) int {
	switch at {
	case nil:
		return -1
	default:
		return at[0]
	}
}

func (p Pattern) Span(command commands.Command) int {
	return spanned(p.expression.FindStringIndex(command.Text))
}

func spanned(at []int) int {
	switch at {
	case nil:
		return 0
	default:
		return at[1] - at[0]
	}
}

func (p Pattern) String() string {
	return p.describes(p.written)
}
