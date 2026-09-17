// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "regexp"

// Leading anchors its expression to the start of a command, ending at a
// word boundary.
type Leading struct {
	described
	opening *regexp.Regexp
	written string
}

func NewLeading(expression, note, description string) (Rule, error) {
	if expression == "" {
		return Leading{described{note, description}, nil, expression}, nil
	}
	opening, err := regexp.Compile(`^(?:` + expression + `)(?:\s|$)`)
	if err != nil {
		return nil, err
	}
	return Leading{described{note, description}, opening, expression}, nil
}

func (l Leading) Matches(command string) bool {
	return l.opening != nil && l.opening.MatchString(command)
}

func (l Leading) String() string {
	return l.describes(l.written)
}
