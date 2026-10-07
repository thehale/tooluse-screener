// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	"github.com/thehale/tooluse-screener/internal/lists"
)

type Branches struct {
	Onto    []string
	NotOnto []string
}

func (b Branches) isUnsaid() bool {
	return len(b.Onto) == 0 && len(b.NotOnto) == 0
}

func (b Branches) isListed(branch string) bool {
	return (len(b.Onto) == 0 || isAmong(b.Onto, branch)) && !isAmong(b.NotOnto, branch)
}

func isAmong(names []string, branch string) bool {
	return lists.Some(names, func(one string) bool { return strings.EqualFold(one, branch) })
}
