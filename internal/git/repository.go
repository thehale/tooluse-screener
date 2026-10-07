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
	branch, unborn := gitOutput(path, "symbolic-ref", "--quiet", "--short", "HEAD")
	listed, unlisted := gitOutput(path, "config", "--list", "-z")
	return repository{branch, settings(listed)}, unborn == nil && unlisted == nil
}

func gitOutput(path string, arguments ...string) (string, error) {
	said, err := exec.Command("git", append([]string{"-C", path}, arguments...)...).Output()
	return strings.TrimSpace(string(said)), err
}

func settings(listed string) map[string]string {
	found := map[string]string{}
	for _, entry := range strings.Split(listed, "\x00") {
		key, value, _ := strings.Cut(entry, "\n")
		found[key] = value
	}
	return found
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
	_, rewritten := r.settings["remote."+remote+".push"]
	branch, known := onto(remote, branchIn(ref))
	return branch, known && !rewritten
}

func (r repository) pushRemote() string {
	return cmp.Or(
		r.settings["branch."+r.branch+".pushremote"],
		r.settings["remote.pushdefault"],
		r.settings["branch."+r.branch+".remote"],
		"origin",
	)
}
