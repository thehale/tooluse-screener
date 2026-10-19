// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"github.com/thehale/tooluse-screener/internal/lists"
	"github.com/thehale/tooluse-screener/internal/shell"
)

func (c Command) TargetDirectories() []string {
	switch {
	case c.Git != nil:
		return c.Git.TargetDirectories()
	case lists.Some(shell.Assignments(c.Text), isRetargetAssignment):
		return []string{""}
	default:
		return nil
	}
}
