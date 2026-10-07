// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

type Wording struct {
	Reason string
	Name   string
}

func (w Wording) nameOr(text string) string {
	if w.Name == "" {
		return text
	} else {
		return w.Name
	}
}
