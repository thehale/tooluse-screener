// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "github.com/thehale/tooluse-screener/internal/rules"

type Policy struct {
	denied  side
	allowed side
}

type side struct {
	commandRules rules.Group
	pathRules    rules.Globs
}
