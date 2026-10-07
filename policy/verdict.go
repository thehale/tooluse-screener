// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "fmt"

type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
	Ask   Decision = "ask"
)

type Verdict struct {
	Decision Decision
	Reason   string
}

func (v Verdict) String() string {
	return fmt.Sprintf("%s: %s", v.Decision, v.Reason)
}
