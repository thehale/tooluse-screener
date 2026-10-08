// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

func nameOr(name, text string) string {
	if name == "" {
		return text
	} else {
		return name
	}
}
