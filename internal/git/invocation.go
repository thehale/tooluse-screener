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
	if words := strings.Fields(commands.WithoutAssignments(command)); len(words) == 0 || words[0] != "git" {
		return Invocation{Unread: unreadNames(command, nil)}
	} else {
		return invocationOf(command, words[1:])
	}
}

func invocationOf(command string, words []string) Invocation {
	directories, global, rest := afterGlobalOptions(append(twinOptions(command), words...))
	subcommand, arguments := subcommandAndArguments(rest)
	return Invocation{
		Git:         true,
		Subcommand:  subcommand,
		Arguments:   arguments,
		Directories: directories,
		Global:      global,
		Unread:      unreadNames(command, twins),
	}
}

func (i Invocation) IsA(subcommand string) bool {
	return i.Subcommand != "" && i.Subcommand == subcommand
}

func (i Invocation) HasAll(words []string) bool {
	return lists.Every(words, i.has)
}

func (i Invocation) has(word string) bool {
	if strings.HasPrefix(word, "-") {
		return slices.Contains(i.Arguments, word)
	} else {
		return len(i.Arguments) > 0 && i.Arguments[0] == word
	}
}

func (i Invocation) Landing(dirs []string) (branch string, isKnown bool) {
	words, isReadable := i.pushWords()
	switch {
	case !isReadable:
		return "", false
	case len(words) == 2 && withoutForce(words[1]) != "HEAD":
		return onto(words[0], branchOf(withoutForce(words[1])))
	default:
		return commonBranch(dirs, func(dir string) (string, bool) { return landingIn(dir, words) })
	}
}

func (i Invocation) pushWords() (words []string, isReadable bool) {
	options, words := optionsAndWords(i.Arguments)
	return words, i.IsA("push") && len(words) <= 2 && len(i.Unread) == 0 && isLandingKept(options, i.Global)
}

func (i Invocation) TargetDirectories() []string {
	if len(i.Unread) > 0 {
		return append(slices.Clone(i.Directories), "")
	} else {
		return i.Directories
	}
}

func afterGlobalOptions(words []string) (directories, global, rest []string) {
	var pointings []pointing
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		option := globalOptionAt(words)
		if isDirectoryOption(option.name) {
			pointings = append(pointings, pointing{option.name, commands.Literal(option.value)})
		}
		global = append(global, option.name)
		words = words[option.width:]
	}
	return directoriesOf(pointings), global, words
}

func isDirectoryOption(option string) bool {
	directoryOptions := []string{"-C", "--git-dir", "--work-tree"}
	return slices.Contains(directoryOptions, option)
}

func subcommandAndArguments(rest []string) (subcommand string, arguments []string) {
	if len(rest) == 0 {
		return "", nil
	} else {
		return rest[0], rest[1:]
	}
}
