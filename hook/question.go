// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"path/filepath"

	"github.com/thehale/tooluse-screener/policy"
)

type Policy interface {
	InDir(directory string) (Checker, error)
}

type Checker interface {
	CheckCommand(line string) policy.Verdict
	CheckPath(path string) policy.Verdict
}

type question func(checker Checker) policy.Verdict

func (call toolCall) posedQuestion() (question, bool) {
	path, isWrite := call.Input.pathFor(call.Tool)
	switch {
	case call.Tool == "Bash":
		return commandQuestion(call.Input.Command), call.Input.Command != ""
	case isWrite:
		return pathQuestion(absolutePath(call.Directory, path)), path != ""
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

func (input toolInput) pathFor(tool string) (string, bool) {
	switch tool {
	case "Write", "Edit", "MultiEdit":
		return input.FilePath, true
	case "NotebookEdit":
		return input.NotebookPath, true
	default:
		return "", false
	}
}

func absolutePath(cwd, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	} else {
		return filepath.Join(cwd, path)
	}
}
