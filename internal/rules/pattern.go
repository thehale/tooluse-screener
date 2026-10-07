// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Pattern struct {
	wording    Wording
	expression *regexp.Regexp
}

func NewPattern(expression string, wording Wording) (Rule, error) {
	regex, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	return Pattern{wording, regex}, nil
}

func (p Pattern) At(command commands.Command) int {
	return start(p.expression.FindStringIndex(command.Text))
}

func start(at []int) int {
	if at == nil {
		return -1
	} else {
		return at[0]
	}
}

func (p Pattern) Span(command commands.Command) int {
	return width(p.expression.FindStringIndex(command.Text))
}

func width(at []int) int {
	if at == nil {
		return 0
	} else {
		return at[1] - at[0]
	}
}

func (p Pattern) Reason(commands.Command) string {
	return p.wording.Reason
}

func (p Pattern) String() string {
	return p.wording.nameOr(p.expression.String())
}
