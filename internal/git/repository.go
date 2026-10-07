// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"cmp"
	"os/exec"
	"strings"

	"github.com/thehale/tooluse-screener/internal/directories"
)

type repository struct {
	branch   string
	settings map[string]string
}

func repositoryAt(dir string) (repository, bool) {
	path := directories.RealPath(dir)
	if path == "" {
		return repository{}, false
	} else {
		return repositoryFrom(path)
	}
}

func repositoryFrom(path string) (repository, bool) {
	branch, branchErr := gitOutput(path, "symbolic-ref", "--quiet", "--short", "HEAD")
	listing, listErr := gitOutput(path, "config", "--list", "-z")
	return repository{branch, settings(listing)}, branchErr == nil && listErr == nil
}

func gitOutput(path string, arguments ...string) (string, error) {
	output, err := exec.Command("git", append([]string{"-C", path}, arguments...)...).Output()
	return strings.TrimSpace(string(output)), err
}

func settings(listing string) map[string]string {
	byKey := map[string]string{}
	for _, entry := range strings.Split(listing, "\x00") {
		key, value, _ := strings.Cut(entry, "\n")
		byKey[key] = value
	}
	return byKey
}

func (r repository) landing(remote string) (string, bool) {
	switch r.settings["push.default"] {
	case "", "simple", "current":
		return r.onto(remote, r.branch)
	case "upstream", "tracking":
		return r.onto(remote, r.settings["branch."+r.branch+".merge"])
	default:
		return "", false
	}
}

func (r repository) onto(remote, ref string) (string, bool) {
	_, isRewritten := r.settings["remote."+remote+".push"]
	branch, isKnown := onto(remote, branchIn(ref))
	return branch, isKnown && !isRewritten
}

func (r repository) pushRemote() string {
	return cmp.Or(
		r.settings["branch."+r.branch+".pushremote"],
		r.settings["remote.pushdefault"],
		r.settings["branch."+r.branch+".remote"],
		"origin",
	)
}
