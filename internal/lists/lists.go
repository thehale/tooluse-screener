// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package lists

import "slices"

func Every[T any](items []T, ok func(T) bool) bool {
	return !slices.ContainsFunc(items, func(item T) bool { return !ok(item) })
}

func Some[T any](items []T, ok func(T) bool) bool {
	return slices.ContainsFunc(items, ok)
}

func Partition[T any](items []T, isFirst func(T) bool) (first, rest []T) {
	for _, item := range items {
		if isFirst(item) {
			first = append(first, item)
		} else {
			rest = append(rest, item)
		}
	}
	return first, rest
}

func Map[T, U any](items []T, change func(T) U) []U {
	changed := make([]U, len(items))
	for index, item := range items {
		changed[index] = change(item)
	}
	return changed
}
