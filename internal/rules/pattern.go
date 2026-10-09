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

func (p Pattern) Find(command commands.Command) (Match, bool) {
	at := p.expression.FindStringIndex(command.Text)
	if at == nil {
		return Match{}, false
	} else {
		return Match{at[0], at[1] - at[0]}, true
	}
}

func (p Pattern) String() string {
	return p.expression.String()
}
