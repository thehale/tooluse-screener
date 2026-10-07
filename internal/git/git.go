// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"slices"
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Invocation struct {
	Git         bool
	Subcommand  string
	Arguments   []string
	Directories []string
	Global      []string
	Unread      []string
}

func Read(command string) Invocation {
	if words := strings.Fields(commands.Spoken(command)); len(words) == 0 || words[0] != "git" {
		return Invocation{Unread: unread(command, nil)}
	} else {
		return invoked(command, words[1:])
	}
}

func invoked(command string, words []string) Invocation {
	directories, global, spoken := afterGlobalOptions(append(twinned(command), words...))
	subcommand, arguments := spokenAs(spoken)
	return Invocation{
		Git:         true,
		Subcommand:  subcommand,
		Arguments:   arguments,
		Directories: directories,
		Global:      global,
		Unread:      unread(command, twins),
	}
}

func (i Invocation) IsA(subcommand string) bool {
	return i.Subcommand != "" && i.Subcommand == subcommand
}

func (i Invocation) Carries(words []string) bool {
	return lists.Every(words, i.carries)
}

func (i Invocation) carries(word string) bool {
	if strings.HasPrefix(word, "-") {
		return slices.Contains(i.Arguments, word)
	} else {
		return len(i.Arguments) > 0 && i.Arguments[0] == word
	}
}

func afterGlobalOptions(words []string) (directories, global, spoken []string) {
	var pointed []pointing
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		option, value, joined := strings.Cut(words[0], "=")
		if pointsAtADirectory(option) {
			pointed = append(pointed, pointing{option, directoryFrom(value, joined, words)})
		}
		global = append(global, option)
		words = words[stride(option, joined, len(words)):]
	}
	return resolved(pointed), global, words
}

func pointsAtADirectory(option string) bool {
	pointing := []string{"-C", "--git-dir", "--work-tree"}
	return slices.Contains(pointing, option)
}

func directoryFrom(value string, joined bool, words []string) string {
	switch {
	case joined:
		return commands.Literal(value)
	case len(words) > 1:
		return commands.Literal(words[1])
	default:
		return ""
	}
}

func stride(option string, joined bool, remaining int) int {
	if joined || !takesAValue(option) || remaining < 2 {
		return 1
	} else {
		return 2
	}
}

func takesAValue(option string) bool {
	withValues := []string{"-C", "-c", "--git-dir", "--namespace", "--work-tree"}
	return slices.Contains(withValues, option)
}

func spokenAs(spoken []string) (subcommand string, arguments []string) {
	if len(spoken) == 0 {
		return "", nil
	} else {
		return spoken[0], spoken[1:]
	}
}
