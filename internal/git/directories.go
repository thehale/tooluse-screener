// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"github.com/thehale/tooluse-screener/internal/directories"
)

type pointing struct {
	option    string
	directory string
}

func directoriesOf(pointings []pointing) []string {
	entries := enteredDirectories(pointings)
	return append(entries, namedDirectories(lastOf(entries), pointings)...)
}

func enteredDirectories(pointings []pointing) []string {
	var entries []string
	here := "."
	for _, one := range pointings {
		if one.option == "-C" {
			here = directories.Within(here, one.directory)
			entries = append(entries, here)
		}
	}
	return entries
}

func namedDirectories(here string, pointings []pointing) []string {
	var dirs []string
	for _, one := range pointings {
		if one.option != "-C" {
			dirs = append(dirs, directories.Within(here, one.directory))
		}
	}
	return dirs
}

func lastOf(dirs []string) string {
	if len(dirs) == 0 {
		return "."
	} else {
		return dirs[len(dirs)-1]
	}
}
