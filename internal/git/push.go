// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"regexp"
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/lists"
)

func Landing(i Invocation) (branch string, known bool) {
	remote, refspec := pushed(i)
	branch = branchOf(refspec)
	return branch, remoteName.MatchString(remote) && branch != ""
}

func pushed(i Invocation) (remote, refspec string) {
	options, words := partitioned(i.Arguments)
	switch {
	case i.IsA("push") && len(words) == 2 && leaveTheLandingAlone(options, i.Global):
		return words[0], words[1]
	default:
		return "", ""
	}
}

func leaveTheLandingAlone(options, global []string) bool {
	return lists.Every(options, changesNeitherEnd) &&
		lists.Every(global, onlyChangesDirectory)
}

func changesNeitherEnd(option string) bool {
	spoken := []string{"-u", "--set-upstream", "-q", "--quiet", "-v", "--verbose", "--progress"}
	return slices.Contains(spoken, option)
}

func onlyChangesDirectory(option string) bool {
	return option == "-C"
}

func partitioned(arguments []string) (options, words []string) {
	for _, argument := range arguments {
		if strings.HasPrefix(argument, "-") {
			options = append(options, argument)
		} else {
			words = append(words, argument)
		}
	}
	return options, words
}

func branchOf(refspec string) string {
	source, destination, paired := strings.Cut(refspec, ":")
	switch {
	case source == "" || strings.HasPrefix(source, "+"):
		return ""
	case !paired:
		return named(source)
	default:
		return named(destination)
	}
}

func named(side string) string {
	branch := strings.TrimPrefix(side, "refs/heads/")
	switch {
	case isABranchName(branch):
		return branch
	default:
		return ""
	}
}

var (
	remoteName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	branchName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
)

func isABranchName(branch string) bool {
	return branchName.MatchString(branch) &&
		!strings.Contains(branch, "..") &&
		!strings.HasPrefix(branch, "refs/")
}
