// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

func twinned(command string) []string {
	var options []string
	for _, assignment := range commands.Assignments(command) {
		name, value, _ := strings.Cut(assignment, "=")
		if option, twin := twins[name]; twin {
			options = append(options, option+"="+value)
		}
	}
	return options
}

var twins = map[string]string{
	"GIT_DIR":       "--git-dir",
	"GIT_WORK_TREE": "--work-tree",
	"GIT_NAMESPACE": "--namespace",
}
