// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Reaching struct {
	rule    Rule
	trusted []string
}

func NewReaching(rule Rule, trusted []string) Rule {
	return Reaching{rule, trusted}
}

func (r Reaching) Matches(command string) bool {
	return r.At(command) >= 0
}

func (r Reaching) At(command string) int {
	switch {
	case r.reaches(command):
		return r.rule.At(command)
	default:
		return -1
	}
}

func (r Reaching) Span(command string) int {
	switch {
	case r.reaches(command):
		return r.rule.Span(command)
	default:
		return 0
	}
}

func (r Reaching) reaches(command string) bool {
	return directories.AllUnder(git.Read(command).Directories, r.trusted)
}

func (r Reaching) Note() string {
	return r.rule.Note()
}

func (r Reaching) String() string {
	return r.rule.String()
}
