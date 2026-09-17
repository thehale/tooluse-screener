// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package lists

import "slices"

func Every[T any](items []T, ok func(T) bool) bool {
	return !slices.ContainsFunc(items, func(item T) bool { return !ok(item) })
}

func Some[T any](items []T, ok func(T) bool) bool {
	return slices.ContainsFunc(items, ok)
}
