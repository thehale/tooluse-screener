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
