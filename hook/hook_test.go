// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/thehale/tooluse-screener/policy"
)

var (
	allow = policy.Verdict{Decision: policy.Allow, Reason: "matches a prefix"}
	deny  = policy.Verdict{Decision: policy.Deny, Reason: "not on my watch"}
	ask   = policy.Verdict{Decision: policy.Ask, Reason: "no opinion"}
)

type always policy.Verdict

func (a always) CheckCommand(string) policy.Verdict {
	return policy.Verdict(a)
}

func (a always) CheckPath(string) policy.Verdict {
	return policy.Verdict(a)
}

type panics struct{}

func (panics) CheckCommand(string) policy.Verdict {
	panic("something went sideways")
}

func (panics) CheckPath(string) policy.Verdict {
	panic("something went sideways")
}

type echoing struct{}

func (echoing) CheckCommand(line string) policy.Verdict {
	return policy.Verdict{Decision: policy.Deny, Reason: line}
}

func (echoing) CheckPath(path string) policy.Verdict {
	return policy.Verdict{Decision: policy.Deny, Reason: path}
}

func writeCall(tool, key, path string) toolCall {
	return toolCall{
		"tool_name":       tool,
		"hook_event_name": "PreToolUse",
		"cwd":             "/work/repo",
		"tool_input":      map[string]any{key: path, "content": "hi"},
	}
}

func claude(command, event string) toolCall {
	return toolCall{
		"tool_name":       "Bash",
		"hook_event_name": event,
		"tool_input":      map[string]any{"command": command},
	}
}

func codex(command, event string) toolCall {
	payload := claude(command, event)
	payload["model"] = "gpt-5.6-sol"
	payload["tool_use_id"] = "exec-0e2a"
	return payload
}

func TestNotOurBusiness(t *testing.T) {
	t.Run("a tool that is not bash", func(t *testing.T) {
		payload := toolCall{"tool_name": "Read", "tool_input": map[string]any{"file_path": "/etc/passwd"}}
		leavesUndecided(t, payload)
	})

	t.Run("a bash call with no command", func(t *testing.T) {
		leavesUndecided(t, toolCall{"tool_name": "Bash", "tool_input": map[string]any{}})
	})

	t.Run("a bash call with an empty command", func(t *testing.T) {
		leavesUndecided(t, claude("", "PreToolUse"))
	})

	t.Run("a payload with no tool at all", func(t *testing.T) {
		leavesUndecided(t, toolCall{})
	})

	t.Run("a tool input that is not an object", func(t *testing.T) {
		for _, given := range []any{nil, []any{}, "ls", 7} {
			leavesUndecided(t, toolCall{"tool_name": "Bash", "tool_input": given})
		}
	})

	t.Run("a command that is not a string", func(t *testing.T) {
		for _, given := range []any{nil, []any{}, map[string]any{}, 7} {
			payload := toolCall{"tool_name": "Bash", "tool_input": map[string]any{"command": given}}
			leavesUndecided(t, payload)
		}
	})

	t.Run("an event the hook does not answer", func(t *testing.T) {
		staysQuiet(t, claude("ls", "PostToolUse"), deny, 0)
	})
}

func TestWhatIsAsked(t *testing.T) {
	t.Run("bash is asked about its command", func(t *testing.T) {
		asks(t, claude("ls -la", "PreToolUse"), "ls -la")
	})

	t.Run("a tool that writes a file is asked about its path", func(t *testing.T) {
		asks(t, writeCall("Write", "file_path", "/work/notes.md"), "/work/notes.md")
		asks(t, writeCall("Edit", "file_path", "/work/notes.md"), "/work/notes.md")
		asks(t, writeCall("MultiEdit", "file_path", "/work/notes.md"), "/work/notes.md")
		asks(t, writeCall("NotebookEdit", "notebook_path", "/work/a.ipynb"), "/work/a.ipynb")
	})

	t.Run("a relative path is read from where the agent is working", func(t *testing.T) {
		asks(t, writeCall("Write", "file_path", "docs/../notes.md"), "/work/repo/notes.md")
	})

	t.Run("a path is read as cleaned", func(t *testing.T) {
		asks(t, writeCall("Write", "file_path", "/work//repo/./notes.md"), "/work/repo/notes.md")
	})

	t.Run("a tool that writes, with no path, is not asked about", func(t *testing.T) {
		leavesUndecided(t, writeCall("Write", "file_path", ""))
		leavesUndecided(t, writeCall("NotebookEdit", "file_path", "/work/a.ipynb"))
	})

	t.Run("a path the policy says nothing about gets no envelope", func(t *testing.T) {
		var out bytes.Buffer
		code := Main(inputOf(writeCall("Write", "file_path", "/work/notes.md")), &out, &bytes.Buffer{}, always(ask))
		if code != 0 || out.String() != "" {
			t.Errorf("exit %d wrote %q, wanted a quiet 0", code, out.String())
		}
	})

	t.Run("a denied path gets a block envelope and a blocking exit", func(t *testing.T) {
		written, code := outcomeOf(t, inputOf(writeCall("Edit", "file_path", "/work/notes.md")), deny)
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		says(t, section(t, envelopeIn(t, written), "hookSpecificOutput"), "permissionDecisionReason", deny.Reason)
	})
}

func asks(t *testing.T, payload toolCall, wanted string) {
	t.Helper()
	if verdict, _ := verdictOn(payload, echoing{}); verdict.Reason != wanted {
		t.Errorf("%v asked about %q, wanted %q", payload, verdict.Reason, wanted)
	}
}

func TestAPolicyThatFails(t *testing.T) {
	t.Run("is treated as ask, said so on stderr, and writes no envelope", func(t *testing.T) {
		var out, complaints bytes.Buffer
		code := Main(inputOf(claude("ls", "PreToolUse")), &out, &complaints, panics{})
		if code != 0 || out.String() != "" {
			t.Errorf("exit %d wrote %q, wanted a quiet 0", code, out.String())
		}
		if !strings.Contains(complaints.String(), "the policy failed") {
			t.Errorf("stderr said %q", complaints.String())
		}
	})
}

func TestWhatEachEventAnswersWith(t *testing.T) {
	answers := []struct {
		event    string
		verdict  policy.Verdict
		envelope string
		code     int
	}{
		{"PreToolUse", allow, "allow", 0},
		{"PreToolUse", deny, "block", 2},
		{"PreToolUse", ask, "", 0},
		{"PermissionRequest", allow, "allow", 0},
		{"PermissionRequest", deny, "deny", 0},
		{"PermissionRequest", ask, "", 0},
	}

	for _, wanted := range answers {
		t.Run(wanted.event+" "+string(wanted.verdict.Decision), func(t *testing.T) {
			payload := claude("git status", wanted.event)
			if wanted.envelope == "" {
				staysQuiet(t, payload, wanted.verdict, wanted.code)
			} else {
				envelope, code := responseTo(t, payload, wanted.verdict)
				speaks(t, envelope, wanted.envelope)
				if code != wanted.code {
					t.Errorf("exit %d, wanted %d", code, wanted.code)
				}
			}

		})
	}
}

func speaks(t *testing.T, envelope map[string]any, wanted string) {
	t.Helper()
	spelled, _ := json.Marshal(envelope)
	if !strings.Contains(string(spelled), `"`+wanted+`"`) {
		t.Errorf("%s does not answer %q", spelled, wanted)
	}
}

func TestClaudePreToolUse(t *testing.T) {
	t.Run("an allowed command gets an allow envelope", func(t *testing.T) {
		envelope, code := responseTo(t, claude("ls", "PreToolUse"), allow)
		if code != 0 {
			t.Errorf("exit %d, wanted 0", code)
		}
		output := section(t, envelope, "hookSpecificOutput")
		says(t, output, "hookEventName", "PreToolUse")
		says(t, output, "permissionDecision", "allow")
		says(t, output, "permissionDecisionReason", allow.Reason)
	})

	t.Run("a denied command gets a block envelope and a blocking exit", func(t *testing.T) {
		envelope, code := responseTo(t, claude("ls", "PreToolUse"), deny)
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		says(t, envelope, "decision", "block")
		says(t, envelope, "reason", deny.Reason)
		says(t, section(t, envelope, "hookSpecificOutput"), "permissionDecision", "deny")
	})

	t.Run("an undecided command gets no envelope so the agent prompts", func(t *testing.T) {
		staysQuiet(t, claude("ls", "PreToolUse"), ask, 0)
	})

	t.Run("the event defaults to PreToolUse when absent", func(t *testing.T) {
		payload := toolCall{"tool_name": "Bash", "tool_input": map[string]any{"command": "ls"}}
		envelope, _ := responseTo(t, payload, deny)
		says(t, section(t, envelope, "hookSpecificOutput"), "hookEventName", "PreToolUse")
	})
}

func TestCodexPreToolUse(t *testing.T) {
	t.Run("enforces a denial", func(t *testing.T) {
		envelope, code := responseTo(t, codex("ls", "PreToolUse"), deny)
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		says(t, section(t, envelope, "hookSpecificOutput"), "permissionDecision", "deny")
	})

	t.Run("stays quiet when undecided", func(t *testing.T) {
		staysQuiet(t, codex("ls", "PreToolUse"), ask, 0)
	})
}

func TestCodexPermissionRequest(t *testing.T) {
	t.Run("allows", func(t *testing.T) {
		envelope, code := responseTo(t, codex("ls", "PermissionRequest"), allow)
		if code != 0 {
			t.Errorf("exit %d, wanted 0", code)
		}
		says(t, decision(t, envelope), "behavior", "allow")
	})

	t.Run("denies and carries the reason as the message", func(t *testing.T) {
		envelope, _ := responseTo(t, codex("ls", "PermissionRequest"), deny)
		says(t, decision(t, envelope), "behavior", "deny")
		says(t, decision(t, envelope), "message", deny.Reason)
	})

	t.Run("stays quiet when undecided", func(t *testing.T) {
		staysQuiet(t, codex("ls", "PermissionRequest"), ask, 0)
	})

	t.Run("never blocks by exit code, because the envelope says everything", func(t *testing.T) {
		for _, verdict := range []policy.Verdict{allow, deny, ask} {
			if _, code := response(codex("ls", "PermissionRequest"), verdict); code != 0 {
				t.Errorf("%s -> exit %d, wanted 0", verdict, code)
			}
		}
	})
}

func TestStdinToStdout(t *testing.T) {
	t.Run("writes the envelope and the blocking exit code", func(t *testing.T) {
		written, code := outcomeOf(t, inputOf(claude("ls", "PreToolUse")), deny)
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		says(t, envelopeIn(t, written), "decision", "block")
	})

	t.Run("a denial carries its reason on stderr, where Codex reads it", func(t *testing.T) {
		var out, complaints bytes.Buffer
		code := Main(inputOf(claude("ls", "PreToolUse")), &out, &complaints, always(deny))
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		if !strings.Contains(complaints.String(), deny.Reason) {
			t.Errorf("stderr said %q, wanted the reason in it", complaints.String())
		}
	})

	t.Run("anything but a denial leaves stderr alone", func(t *testing.T) {
		for _, verdict := range []policy.Verdict{allow, ask} {
			var out, complaints bytes.Buffer
			Main(inputOf(claude("ls", "PreToolUse")), &out, &complaints, always(verdict))
			if complaints.String() != "" {
				t.Errorf("%s wrote %q to stderr", verdict.Decision, complaints.String())
			}
		}
	})

	t.Run("writes nothing when undecided", func(t *testing.T) {
		writesNothing(t, inputOf(claude("ls", "PreToolUse")), ask)
	})

	t.Run("unreadable input is not something to have an opinion about", func(t *testing.T) {
		writesNothing(t, strings.NewReader("not json at all"), deny)
	})

	t.Run("a JSON payload that is not an object is ignored", func(t *testing.T) {
		writesNothing(t, strings.NewReader("[1, 2, 3]"), deny)
	})

	t.Run("a payload with anything after it is ignored", func(t *testing.T) {
		writesNothing(t, inputWithTrailer("ls", " trailing"), deny)
	})

	t.Run("a payload with only blank space after it is answered", func(t *testing.T) {
		written, code := outcomeOf(t, inputWithTrailer("ls", "  \n\n"), deny)
		if code != 2 {
			t.Errorf("exit %d wrote %q, wanted 2", code, written)
		}
	})

	t.Run("a null payload is ignored", func(t *testing.T) {
		writesNothing(t, strings.NewReader("null"), deny)
	})
}

func leavesUndecided(t *testing.T, payload toolCall) {
	t.Helper()
	if verdict, asked := verdictOn(payload, always(deny)); asked {
		t.Errorf("%v -> %s, wanted no opinion", payload, verdict)
	}
}

func responseTo(t *testing.T, payload toolCall, verdict policy.Verdict) (map[string]any, int) {
	t.Helper()
	envelope, code := response(payload, verdict)
	if envelope == nil {
		t.Fatalf("%v -> nothing, wanted an envelope", payload)
	}
	return envelope, code
}

func staysQuiet(t *testing.T, payload toolCall, verdict policy.Verdict, wanted int) {
	t.Helper()
	envelope, code := response(payload, verdict)
	if envelope != nil || code != wanted {
		t.Errorf("%v -> %v, exit %d, wanted nothing and %d", payload, envelope, code, wanted)
	}
}

func section(t *testing.T, envelope map[string]any, key string) map[string]any {
	t.Helper()
	nested, is := envelope[key].(map[string]any)
	if !is {
		t.Fatalf("%v has no %s", envelope, key)
	}
	return nested
}

func decision(t *testing.T, envelope map[string]any) map[string]any {
	t.Helper()
	return section(t, section(t, envelope, "hookSpecificOutput"), "decision")
}

func says(t *testing.T, envelope map[string]any, key, wanted string) {
	t.Helper()
	if said, _ := envelope[key].(string); said != wanted {
		t.Errorf("%s = %q, wanted %q", key, said, wanted)
	}
}

func inputWithTrailer(command, after string) io.Reader {
	written, _ := json.Marshal(claude(command, "PreToolUse"))
	return strings.NewReader(string(written) + after)
}

func inputOf(payload toolCall) *bytes.Reader {
	written, _ := json.Marshal(payload)
	return bytes.NewReader(written)
}

func outcomeOf(t *testing.T, in io.Reader, verdict policy.Verdict) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := Main(in, &out, &bytes.Buffer{}, always(verdict))
	return out.String(), code
}

func envelopeIn(t *testing.T, written string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(written), &envelope); err != nil {
		t.Fatalf("%q: %v", written, err)
	}
	return envelope
}

func writesNothing(t *testing.T, in io.Reader, verdict policy.Verdict) {
	t.Helper()
	if written, code := outcomeOf(t, in, verdict); written != "" || code != 0 {
		t.Errorf("wrote %q and exited %d, wanted nothing and 0", written, code)
	}
}
