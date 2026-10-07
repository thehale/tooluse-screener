// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Rule interface {
	At(command commands.Command) int
	Span(command commands.Command) int
	Reason(command commands.Command) string
	String() string
}

type wording struct {
	reason string
	name   string
}

func (w wording) Reason(commands.Command) string {
	return w.reason
}

func (w wording) nameOr(text string) string {
	if w.name == "" {
		return text
	} else {
		return w.name
	}
}
