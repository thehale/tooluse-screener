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

func resolved(pointed []pointing) []string {
	entered := enteredIn(pointed)
	return append(entered, namedFrom(lastOf(entered), pointed)...)
}

func enteredIn(pointed []pointing) []string {
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

func namedFrom(here string, pointed []pointing) []string {
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

func (i Invocation) PointedAt() []string {
	if len(i.Unread) > 0 {
		return append(slices.Clone(i.Directories), "")
	} else {
		return i.Directories
	}
}
