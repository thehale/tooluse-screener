// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"slices"

	"github.com/thehale/tooluse-screener/internal/directories"
)

type pointing struct {
	option    string
	directory string
}

func directoriesOf(pointed []pointing) []string {
	entered := enteredDirectories(pointed)
	return append(entered, namedDirectories(lastOf(entered), pointed)...)
}

func enteredDirectories(pointed []pointing) []string {
	var entered []string
	here := "."
	for _, one := range pointed {
		if one.option == "-C" {
			here = directories.Within(here, one.directory)
			entered = append(entered, here)
		}
	}
	return entered
}

func namedDirectories(here string, pointed []pointing) []string {
	var named []string
	for _, one := range pointed {
		if one.option != "-C" {
			named = append(named, directories.Within(here, one.directory))
		}
	}
	return named
}

func lastOf(entered []string) string {
	if len(entered) == 0 {
		return "."
	} else {
		return entered[len(entered)-1]
	}
}

func (i Invocation) TargetDirectories() []string {
	if len(i.Unread) > 0 {
		return append(slices.Clone(i.Directories), "")
	} else {
		return i.Directories
	}
}
