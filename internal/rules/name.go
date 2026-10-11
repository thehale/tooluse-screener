// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

func nameOr(name, text string) string {
	if name == "" {
		return text
	} else {
		return name
	}
}
