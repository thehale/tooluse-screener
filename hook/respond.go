// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import "github.com/thehale/tooluse-screener/policy"

func respond(payload toolCall, verdict policy.Verdict) (envelope map[string]any, code int) {
	switch event(payload) {
	case "PermissionRequest":
		return permissionRequest(verdict), 0
	case "PreToolUse":
		return preToolUse(verdict)
	default:
		return nil, 0
	}
}

const blocked = 2

func event(payload toolCall) string {
	if named := text(payload, "hook_event_name"); named == "" {
		return "PreToolUse"
	} else {
		return named
	}
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
