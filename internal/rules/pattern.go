// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"errors"
	"fmt"
	"regexp"
	"regexp/syntax"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Pattern struct {
	expression *regexp.Regexp
}

func NewPattern(expression string) (Pattern, error) {
	regex, err := regexp.Compile(expression)
	if err != nil {
		return Pattern{}, fmt.Errorf("invalid regular expression `%s`: %s", expression, syntaxMessage(err))
	} else {
		return Pattern{regex}, nil
	}
}

func syntaxMessage(err error) string {
	var syntaxError *syntax.Error
	if errors.As(err, &syntaxError) {
		return fmt.Sprintf("%s: `%s`", syntaxError.Code, syntaxError.Expr)
	} else {
		return err.Error()
	}
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

func (p Pattern) String() string {
	return p.expression.String()
}
