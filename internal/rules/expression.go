// Copyright (c) 2026 Joseph Hale
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
