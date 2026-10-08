// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Expression interface {
	At(command commands.Command) int
	Span(command commands.Command) int
	String() string
}
