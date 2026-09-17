// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "regexp"

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

func (p Pattern) Matches(command string) bool {
	return p.expression.MatchString(command)
}

func (p Pattern) String() string {
	return p.describes(p.written)
}
