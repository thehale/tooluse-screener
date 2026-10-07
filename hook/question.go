// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"path/filepath"

	"github.com/thehale/tooluse-screener/policy"
)

type Checker interface {
	CheckCommand(line string) policy.Verdict
	CheckPath(path string) (policy.Verdict, bool)
}

func Decide(payload Payload, checker Checker) (policy.Verdict, bool) {
	if asked, posed := questionIn(payload); !posed {
		return policy.Verdict{}, false
	} else {
		return asked(checker)
	}
}

type question func(checker Checker) (policy.Verdict, bool)

func questionIn(payload Payload) (question, bool) {
	arguments, given := payload["tool_input"].(map[string]any)
	tool := text(payload, "tool_name")
	target, writes := targets[tool]
	switch {
	case !given:
		return nil, false
	case tool == "Bash":
		command := text(Payload(arguments), "command")
		return aboutCommand(command), command != ""
	case writes:
		path := text(Payload(arguments), target)
		return aboutPath(resolved(text(payload, "cwd"), path)), path != ""
	default:
		return nil, false
	}
}

func aboutCommand(line string) question {
	return func(checker Checker) (policy.Verdict, bool) {
		return checker.CheckCommand(line), true
	}
}

func aboutPath(path string) question {
	return func(checker Checker) (policy.Verdict, bool) {
		return checker.CheckPath(path)
	}
}

var targets = map[string]string{
	"Write":        "file_path",
	"Edit":         "file_path",
	"MultiEdit":    "file_path",
	"NotebookEdit": "notebook_path",
}

func resolved(cwd, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	} else {
		return filepath.Join(cwd, path)
	}
}
