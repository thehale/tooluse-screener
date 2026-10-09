// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Match struct {
	Start, Width int
}

type Expression interface {
	Find(command commands.Command) (Match, bool)
	String() string
}
