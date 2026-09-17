// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

// Package hook answers an agent before it runs a Bash command.
//
// It reads the agent's hook payload from stdin and writes back the
// envelope that agent expects for that event:
//
//	| Agent  | Event               | allow          | deny           | ask       |
//	| ------ | ------------------- | -------------- | -------------- | --------- |
//	| Claude | `PreToolUse`        | allow envelope | block envelope | no output |
//	| Codex  | `PreToolUse`        | no output      | block envelope | no output |
//	| Codex  | `PermissionRequest` | allow envelope | deny envelope  | no output |
//
// No output means "no opinion", and the agent prompts as it normally
// would. Codex only enforces a denial on `PreToolUse`, so it calls this
// on both events.
package hook

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/thehale/tooluse-screener/internal/check"
)

type Payload map[string]any

type Checking func(line string) check.Verdict

// Decide returns false for a payload holding no Bash command.
func Decide(payload Payload, checking Checking) (check.Verdict, bool) {
	command, asked := bashCommandIn(payload)
	if !asked {
		return check.Verdict{}, false
	}
	return checking(command), true
}

// Respond returns a nil envelope where the agent is to be told nothing.
func Respond(payload Payload, verdict check.Verdict) (envelope map[string]any, code int) {
	switch event(payload) {
	case "PermissionRequest":
		return permissionRequest(verdict), 0
	case "PreToolUse":
		return preToolUse(verdict, isCodex(payload))
	default:
		return nil, 0
	}
}

func Main(in io.Reader, out, complaints io.Writer, checking Checking) int {
	payload, read := payloadOn(in)
	if !read {
		return 0
	}
	verdict, asked := Decide(payload, noting(checking, complaints))
	if !asked {
		return 0
	}
	envelope, code := Respond(payload, verdict)
	if envelope != nil {
		written, _ := json.Marshal(envelope)
		_, _ = fmt.Fprint(out, string(written))
	}
	return code
}

var failOpen = check.Verdict{Decision: check.Ask, Reason: "The policy could not answer"}

const blocked = 2

func bashCommandIn(payload Payload) (string, bool) {
	if text(payload, "tool_name") != "Bash" {
		return "", false
	}
	arguments, given := payload["tool_input"].(map[string]any)
	if !given {
		return "", false
	}
	command := text(Payload(arguments), "command")
	return command, command != ""
}

func noting(checking Checking, complaints io.Writer) Checking {
	return func(line string) (verdict check.Verdict) {
		defer func() {
			if failure := recover(); failure != nil {
				_, _ = fmt.Fprintf(complaints, "tooluse-screener: the policy failed: %v\n", failure)
				verdict = failOpen
			}
		}()
		return checking(line)
	}
}

func event(payload Payload) string {
	if named := text(payload, "hook_event_name"); named != "" {
		return named
	}
	return "PreToolUse"
}

func isCodex(payload Payload) bool {
	return text(payload, "turn_id") != ""
}

func preToolUse(verdict check.Verdict, isCodex bool) (map[string]any, int) {
	switch {
	case verdict.Decision == check.Deny:
		return map[string]any{
			"decision": "block",
			"reason":   verdict.Reason,
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": verdict.Reason,
			},
		}, blocked
	case verdict.Decision == check.Allow && !isCodex:
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

func permissionRequest(verdict check.Verdict) map[string]any {
	var decision map[string]any
	switch verdict.Decision {
	case check.Deny:
		decision = map[string]any{"behavior": "deny", "message": verdict.Reason}
	case check.Allow:
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
