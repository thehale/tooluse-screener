// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"path/filepath"

	"github.com/thehale/tooluse-screener/policy"
)

type Checker interface {
	CheckCommand(line string) policy.Verdict
	CheckPath(path string) policy.Verdict
}

func (call toolCall) verdict(checker Checker) (policy.Verdict, bool) {
	if query, isPosed := call.posedQuestion(); !isPosed {
		return policy.Verdict{}, false
	} else {
		return query(checker), true
	}
}

type question func(checker Checker) policy.Verdict

func (call toolCall) posedQuestion() (question, bool) {
	arguments, hasInput := call["tool_input"].(map[string]any)
	tool := call.text("tool_name")
	target, isWrite := targets[tool]
	switch {
	case !hasInput:
		return nil, false
	case tool == "Bash":
		command := toolCall(arguments).text("command")
		return commandQuestion(command), command != ""
	case isWrite:
		path := toolCall(arguments).text(target)
		return pathQuestion(absolutePath(call.text("cwd"), path)), path != ""
	default:
		return nil, false
	}
}

func commandQuestion(line string) question {
	return func(checker Checker) policy.Verdict {
		return checker.CheckCommand(line)
	}
}

func pathQuestion(path string) question {
	return func(checker Checker) policy.Verdict {
		return checker.CheckPath(path)
	}
}

var targets = map[string]string{
	"Write":        "file_path",
	"Edit":         "file_path",
	"MultiEdit":    "file_path",
	"NotebookEdit": "notebook_path",
}

func absolutePath(cwd, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	} else {
		return filepath.Join(cwd, path)
	}
}
