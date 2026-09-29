// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package check

import "github.com/thehale/tooluse-screener/internal/policy"

type Question struct {
	Command string
	Path    string
}

func Answer(asked Question, p policy.Policy) (Verdict, bool) {
	switch {
	case asked.Path != "":
		return Writing(asked.Path, p)
	default:
		return Evaluate(asked.Command, p), true
	}
}
