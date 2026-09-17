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
	Subcommand  string
	Arguments   []string
	Directories []string
	Global      []string
}

func Read(command string) Invocation {
	words := strings.Fields(commands.Spoken(command))
	if len(words) == 0 || words[0] != "git" {
		return Invocation{}
	}
	directories, global, spoken := afterGlobalOptions(words[1:])
	return spokenAs(spoken, directories, global)
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
	}
	return len(i.Arguments) > 0 && i.Arguments[0] == word
}

func afterGlobalOptions(words []string) (directories, global, spoken []string) {
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		option, value, joined := strings.Cut(words[0], "=")
		if pointsAtADirectory(option) {
			directories = append(directories, directoryFrom(value, joined, words))
		}
		global = append(global, option)
		words = words[stride(option, joined, len(words)):]
	}
	return directories, global, words
}

func pointsAtADirectory(option string) bool {
	pointing := []string{"-C", "--git-dir", "--work-tree"}
	return slices.Contains(pointing, option)
}

func directoryFrom(value string, joined bool, words []string) string {
	switch {
	case joined:
		return value
	case len(words) > 1:
		return words[1]
	default:
		return ""
	}
}

func stride(option string, joined bool, remaining int) int {
	if joined || !takesAValue(option) || remaining < 2 {
		return 1
	}
	return 2
}

func takesAValue(option string) bool {
	withValues := []string{"-C", "-c", "--git-dir", "--namespace", "--work-tree"}
	return slices.Contains(withValues, option)
}

func spokenAs(spoken, directories, global []string) Invocation {
	if len(spoken) == 0 {
		return Invocation{Directories: directories, Global: global}
	}
	return Invocation{
		Subcommand:  spoken[0],
		Arguments:   spoken[1:],
		Directories: directories,
		Global:      global,
	}
}
