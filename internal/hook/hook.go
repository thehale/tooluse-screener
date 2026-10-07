// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/thehale/tooluse-screener/internal/policy"
)

type Payload map[string]any

type Checker interface {
	CheckCommand(line string) policy.Verdict
	CheckPath(path string) (policy.Verdict, bool)
}

func Decide(payload Payload, checker Checker) (policy.Verdict, bool) {
	asked, posed := questionIn(payload)
	if !posed {
		return policy.Verdict{}, false
	}
	return asked(checker)
}

func Respond(payload Payload, verdict policy.Verdict) (envelope map[string]any, code int) {
	switch event(payload) {
	case "PermissionRequest":
		return permissionRequest(verdict), 0
	case "PreToolUse":
		return preToolUse(verdict)
	default:
		return nil, 0
	}
}

func Main(in io.Reader, out, complaints io.Writer, checker Checker) int {
	payload, read := payloadOn(in)
	if !read {
		return 0
	}
	verdict, asked := decided(payload, checker, complaints)
	if !asked {
		return 0
	}
	envelope, code := Respond(payload, verdict)
	wrote(envelope, out)
	if code == blocked {
		_, _ = fmt.Fprintln(complaints, verdict.Reason)
	}
	return code
}

func wrote(envelope map[string]any, out io.Writer) {
	if envelope != nil {
		written, _ := json.Marshal(envelope)
		_, _ = fmt.Fprint(out, string(written))
	}
}

var failOpen = policy.Verdict{Decision: policy.Ask, Reason: "The policy could not answer"}

const blocked = 2

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
	switch {
	case filepath.IsAbs(path):
		return filepath.Clean(path)
	default:
		return filepath.Join(cwd, path)
	}
}

func decided(payload Payload, checker Checker, complaints io.Writer) (verdict policy.Verdict, asked bool) {
	defer func() {
		if failure := recover(); failure != nil {
			_, _ = fmt.Fprintf(complaints, "tooluse-screener: the policy failed: %v\n", failure)
			verdict, asked = failOpen, true
		}
	}()
	return Decide(payload, checker)
}

func event(payload Payload) string {
	if named := text(payload, "hook_event_name"); named != "" {
		return named
	}
	return "PreToolUse"
}

func preToolUse(verdict policy.Verdict) (map[string]any, int) {
	switch verdict.Decision {
	case policy.Deny:
		return map[string]any{
			"decision": "block",
			"reason":   verdict.Reason,
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": verdict.Reason,
			},
		}, blocked
	case policy.Allow:
		return map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "allow",
				"permissionDecisionReason": verdict.Reason,
			},
		}, 0
	default:
		return nil, 0
	}
}

func permissionRequest(verdict policy.Verdict) map[string]any {
	var decision map[string]any
	switch verdict.Decision {
	case policy.Deny:
		decision = map[string]any{"behavior": "deny", "message": verdict.Reason}
	case policy.Allow:
		decision = map[string]any{"behavior": "allow"}
	default:
		return nil
	}
	return map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName": "PermissionRequest",
			"decision":      decision,
		},
	}
}

func payloadOn(in io.Reader) (Payload, bool) {
	var payload map[string]any
	reading := json.NewDecoder(in)
	if err := reading.Decode(&payload); err != nil {
		return nil, false
	}
	if _, err := reading.Token(); err != io.EOF {
		return nil, false
	}
	return payload, payload != nil
}

func text(payload Payload, key string) string {
	written, _ := payload[key].(string)
	return written
}
