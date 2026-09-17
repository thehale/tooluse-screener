// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

type Rule interface {
	Matches(command string) bool
	Note() string
	String() string
}

type described struct {
	note        string
	description string
}

func (d described) Note() string {
	return d.note
}

func (d described) describes(written string) string {
	if d.description == "" {
		return written
	}
	return d.description
}
