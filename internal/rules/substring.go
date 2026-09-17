// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "strings"

type Substring struct {
	described
	text string
}

func NewSubstring(text, note, description string) Rule {
	return Substring{described{note, description}, text}
}

func (s Substring) Matches(command string) bool {
	return s.text != "" && strings.Contains(command, s.text)
}

func (s Substring) String() string {
	return s.describes(s.text)
}
