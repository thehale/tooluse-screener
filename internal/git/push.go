// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"cmp"
	"regexp"
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/lists"
)

func isForceOption(option string) bool {
	forceOptions := []string{"-f", "--force", "--force-with-lease", "--force-if-includes"}
	return slices.Contains(forceOptions, option) || strings.HasPrefix(option, "--force-with-lease=")
}

func withoutForce(refspec string) string {
	return strings.TrimPrefix(refspec, "+")
}

func onto(remote, branch string) (string, bool) {
	return branch, remoteName.MatchString(remote) && branch != ""
}

func commonBranch(dirs []string, landing func(string) (string, bool)) (string, bool) {
	branches := make([]string, 0, len(dirs))
	isKnown := true
	for _, dir := range dirs {
		branch, isRead := landing(dir)
		branches = append(branches, branch)
		isKnown = isKnown && isRead
	}
	slices.Sort(branches)
	distinct := slices.Compact(branches)
	return cmp.Or(distinct...), isKnown && len(distinct) == 1
}

func landingIn(dir string, words []string) (string, bool) {
	repo, isOpen := repositoryAt(dir)
	switch {
	case !isOpen:
		return "", false
	case len(words) == 2:
		return onto(words[0], branchIn(repo.branch))
	case len(words) == 1:
		return repo.landing(words[0])
	default:
		return repo.landing(repo.pushRemote())
	}
}

func isLandingKept(options, global []string) bool {
	return lists.Every(options, isLandingKeptBy) &&
		lists.Every(global, isDirectoryOnly)
}

func isLandingKeptBy(option string) bool {
	neutralOptions := []string{"-u", "--set-upstream", "-q", "--quiet", "-v", "--verbose", "--progress"}
	return slices.Contains(neutralOptions, option) || isForceOption(option)
}

func isDirectoryOnly(option string) bool {
	return option == "-C"
}

func optionsAndWords(arguments []string) (options, words []string) {
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
	source, destination, isPaired := strings.Cut(refspec, ":")
	switch {
	case source == "":
		return ""
	case !isPaired:
		return branchIn(source)
	default:
		return branchIn(destination)
	}
}

func branchIn(side string) string {
	branch := strings.TrimPrefix(side, "refs/heads/")
	if isABranchName(branch) {
		return branch
	} else {
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
