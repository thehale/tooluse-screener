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

func unlistedVerdict(subject string) Verdict {
	return Verdict{Ask, fmt.Sprintf("%s is not in the shared allow list", subject)}
}

func denyReason(subject string, rule fmt.Stringer, reason string) string {
	message := fmt.Sprintf("%s matches a denied rule: %s", subject, rule)
	if reason == "" {
		return message
	} else {
		return fmt.Sprintf("%s. %s", message, reason)
	}
}
