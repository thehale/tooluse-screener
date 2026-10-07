// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/lists"
)

func movesAfter(command string, moved []string) []string {
	words := strings.Fields(WithoutAssignments(command))
	switch {
	case len(words) == 0:
		return retargetsAfter(Assignments(command), moved)
	case slices.Contains(exporting, words[0]):
		return retargetsAfter(words[1:], moved)
	case words[0] == "cd":
		return arrivals(moved, destination(words[1:], "~"))
	case words[0] == "pushd":
		return arrivals(moved, destination(words[1:], ""))
	case slices.Contains(unseenMoves, words[0]):
		return arrivals(moved, "")
	default:
		return moved
	}
}

var unseenMoves = []string{"popd", "source", ".", "eval"}

var exporting = []string{"export", "declare", "typeset", "readonly", "local"}

func retargetsAfter(words []string, moved []string) []string {
	if lists.Some(words, isRetargetAssignment) {
		return arrivals(moved, "")
	} else {
		return moved
	}
}

func isRetargetAssignment(word string) bool {
	name, _, _ := strings.Cut(word, "=")
	return IsRetargeting(name)
}

func destination(arguments []string, bare string) string {
	switch {
	case len(arguments) == 0:
		return bare
	case len(arguments) == 2 && arguments[0] == "--":
		return Literal(arguments[1])
	case len(arguments) == 1 && !strings.ContainsRune("-+", rune(arguments[0][0])):
		return Literal(arguments[0])
	default:
		return ""
	}
}

func arrivals(moved []string, destination string) []string {
	arrived := slices.Clone(moved)
	for _, from := range append([]string{"."}, moved...) {
		arrived = append(arrived, directories.Destination(from, destination))
	}
	slices.Sort(arrived)
	return slices.DeleteFunc(slices.Compact(arrived), isHere)
}

func isHere(directory string) bool {
	return directory == "."
}
