// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Pattern struct {
	named
	expression *regexp.Regexp
	written    string
}

func NewPattern(expression, reason, name string) (Rule, error) {
	built, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	return Pattern{named{reason, name}, built, expression}, nil
}

func (p Pattern) At(command commands.Command) int {
	return opened(p.expression.FindStringIndex(command.Text))
}

func opened(at []int) int {
	if at == nil {
		return -1
	} else {
		return at[0]
	}
}

func (p Pattern) Span(command commands.Command) int {
	return spanned(p.expression.FindStringIndex(command.Text))
}

func spanned(at []int) int {
	if at == nil {
		return 0
	} else {
		return at[1] - at[0]
	}
}

func (p Pattern) String() string {
	return p.calls(p.written)
}
