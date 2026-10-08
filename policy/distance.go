// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

func editDistance(from, to string) int {
	previous := make([]int, len(to)+1)
	for index := range previous {
		previous[index] = index
	}
	for row := 1; row <= len(from); row++ {
		previous = nextRow(previous, from[row-1], to, row)
	}
	return previous[len(to)]
}

func nextRow(previous []int, letter byte, to string, row int) []int {
	current := make([]int, len(to)+1)
	current[0] = row
	for column := 1; column <= len(to); column++ {
		current[column] = min(previous[column]+1, current[column-1]+1, previous[column-1]+substitution(letter, to[column-1]))
	}
	return current
}

func substitution(from, to byte) int {
	if from == to {
		return 0
	} else {
		return 1
	}
}
