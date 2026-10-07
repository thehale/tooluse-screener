// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/lists"
)

func movesAfter(command string, destinations []string) []string {
	words := strings.Fields(WithoutAssignments(command))
	switch {
	case len(words) == 0:
		return retargetsAfter(Assignments(command), destinations)
	case slices.Contains(exporting, words[0]):
		return retargetsAfter(words[1:], destinations)
	case words[0] == "cd":
		return arrivals(destinations, destination(words[1:], "~"))
	case words[0] == "pushd":
		return arrivals(destinations, destination(words[1:], ""))
	case slices.Contains(unseenMoves, words[0]):
		return arrivals(destinations, "")
	default:
		return destinations
	}
}

var unseenMoves = []string{"popd", "source", ".", "eval"}

var exporting = []string{"export", "declare", "typeset", "readonly", "local"}

func retargetsAfter(words []string, destinations []string) []string {
	if lists.Some(words, isRetargetAssignment) {
		return arrivals(destinations, "")
	} else {
		return destinations
	}
}

func isRetargetAssignment(word string) bool {
	name, _, _ := strings.Cut(word, "=")
	return IsRetargeting(name)
}

func destination(arguments []string, bareDestination string) string {
	switch {
	case len(arguments) == 0:
		return bareDestination
	case len(arguments) == 2 && arguments[0] == "--":
		return Literal(arguments[1])
	case len(arguments) == 1 && !strings.ContainsRune("-+", rune(arguments[0][0])):
		return Literal(arguments[0])
	default:
		return ""
	}
}

func arrivals(earlier []string, destination string) []string {
	destinations := slices.Clone(earlier)
	for _, from := range append([]string{"."}, earlier...) {
		destinations = append(destinations, directories.Destination(from, destination))
	}
	slices.Sort(destinations)
	return slices.DeleteFunc(slices.Compact(destinations), isHere)
}

func isHere(directory string) bool {
	return directory == "."
}
