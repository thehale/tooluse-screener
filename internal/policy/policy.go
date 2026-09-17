// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "github.com/thehale/tooluse-screener/internal/rules"

type Policy struct {
	Denied  rules.Group
	Allowed rules.Group
}
