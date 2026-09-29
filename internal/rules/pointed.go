// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "github.com/thehale/tooluse-screener/internal/git"

func pointedAt(invocation git.Invocation) []string {
	switch {
	case len(invocation.Unread) > 0:
		return append(invocation.Directories, "")
	default:
		return invocation.Directories
	}
}
