// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "strings"

// Prefix ends at a word boundary: `ls` matches `ls -la`, not `lsblk`.
type Prefix struct {
	described
	text string
}

func NewPrefix(text, note, description string) Rule {
	return Prefix{described{note, description}, text}
}

func (p Prefix) Matches(command string) bool {
	return p.text != "" &&
		(command == p.text || strings.HasPrefix(command, p.text+" "))
}

func (p Prefix) String() string {
	return p.describes(p.text)
}
