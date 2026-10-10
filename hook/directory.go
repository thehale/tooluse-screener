// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"os"

	"github.com/thehale/tooluse-screener/policy"
)

func inDirectory(cwd string, query question) question {
	if cwd == "" || os.Chdir(cwd) == nil {
		return query
	} else {
		return lost(cwd)
	}
}

func lost(cwd string) question {
	return func(Checker) policy.Verdict {
		return policy.Verdict{Decision: policy.Ask, Reason: "The session directory " + cwd + " could not be entered"}
	}
}
