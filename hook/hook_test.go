// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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

type wherever struct{}

func (wherever) CheckCommand(string) policy.Verdict {
	return policy.Verdict{Decision: policy.Deny, Reason: here()}
}

func (wherever) CheckPath(string) policy.Verdict {
	return policy.Verdict{Decision: policy.Deny, Reason: here()}
}

func here() string {
	wd, _ := os.Getwd()
	return wd
}

func writeCall(t *testing.T, tool, key, path string) toolCall {
	return toolCall{
		"tool_name":       tool,
		"hook_event_name": "PreToolUse",
		"cwd":             t.TempDir(),
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
		for _, input := range []any{nil, []any{}, "ls", 7} {
			leavesUndecided(t, toolCall{"tool_name": "Bash", "tool_input": input})
		}
	})

	t.Run("a command that is not a string", func(t *testing.T) {
		for _, input := range []any{nil, []any{}, map[string]any{}, 7} {
			payload := toolCall{"tool_name": "Bash", "tool_input": map[string]any{"command": input}}
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
		asks(t, writeCall(t, "Write", "file_path", "/work/notes.md"), "/work/notes.md")
		asks(t, writeCall(t, "Edit", "file_path", "/work/notes.md"), "/work/notes.md")
		asks(t, writeCall(t, "MultiEdit", "file_path", "/work/notes.md"), "/work/notes.md")
		asks(t, writeCall(t, "NotebookEdit", "notebook_path", "/work/a.ipynb"), "/work/a.ipynb")
	})

	t.Run("a relative path is read from where the agent is working", func(t *testing.T) {
		call := writeCall(t, "Write", "file_path", "docs/../notes.md")
		asks(t, call, filepath.Join(call.text("cwd"), "notes.md"))
	})

	t.Run("a path is read as cleaned", func(t *testing.T) {
		asks(t, writeCall(t, "Write", "file_path", "/work//repo/./notes.md"), "/work/repo/notes.md")
	})

	t.Run("a tool that writes, with no path, is not asked about", func(t *testing.T) {
		leavesUndecided(t, writeCall(t, "Write", "file_path", ""))
		leavesUndecided(t, writeCall(t, "NotebookEdit", "file_path", "/work/a.ipynb"))
	})

	t.Run("a path the policy says nothing about gets no envelope", func(t *testing.T) {
		var out bytes.Buffer
		code := Main(inputOf(writeCall(t, "Write", "file_path", "/work/notes.md")), &out, &bytes.Buffer{}, always(ask))
		if code != 0 || out.String() != "" {
			t.Errorf("exit %d wrote %q, wanted a quiet 0", code, out.String())
		}
	})

	t.Run("a denied path gets a block envelope and a blocking exit", func(t *testing.T) {
		output, code := outcomeOf(t, inputOf(writeCall(t, "Edit", "file_path", "/work/notes.md")), deny)
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		says(t, section(t, envelopeIn(t, output), "hookSpecificOutput"), "permissionDecisionReason", deny.Reason)
	})
}

func asks(t *testing.T, payload toolCall, wanted string) {
	t.Helper()
	if verdict, _ := payload.verdict(echoing{}); verdict.Reason != wanted {
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

func TestJudgingFromTheSessionDirectory(t *testing.T) {
	t.Run("a bash call is judged from the payload's cwd, not the process's", func(t *testing.T) {
		t.Chdir(t.TempDir())
		there := t.TempDir()
		payload := claude("ls", "PreToolUse")
		payload["cwd"] = there
		judgedAt(t, payload, there)
	})

	t.Run("no cwd still judges from the process's directory", func(t *testing.T) {
		t.Chdir(t.TempDir())
		judgedAt(t, claude("ls", "PreToolUse"), here())
	})

	t.Run("a cwd that does not exist asks, without consulting the checker", func(t *testing.T) {
		payload := claude("ls", "PreToolUse")
		payload["cwd"] = filepath.Join(t.TempDir(), "gone")
		var out, complaints bytes.Buffer
		code := Main(inputOf(payload), &out, &complaints, panics{})
		if code != 0 || out.String() != "" || complaints.String() != "" {
			t.Errorf("exit %d wrote %q and %q, wanted a quiet ask", code, out.String(), complaints.String())
		}
	})

	t.Run("a cwd that does not exist leaves a call it does not judge alone", func(t *testing.T) {
		payload := toolCall{"tool_name": "Read", "cwd": filepath.Join(t.TempDir(), "gone"), "tool_input": map[string]any{"file_path": "a"}}
		leavesUndecided(t, payload)
	})
}

func judgedAt(t *testing.T, payload toolCall, dir string) {
	t.Helper()
	var out, complaints bytes.Buffer
	Main(inputOf(payload), &out, &complaints, wherever{})
	output := section(t, envelopeIn(t, out.String()), "hookSpecificOutput")
	reason, _ := output["permissionDecisionReason"].(string)
	if resolved(reason) != resolved(dir) {
		t.Errorf("judged from %q, wanted %q", reason, dir)
	}
}

func resolved(path string) string {
	target, _ := filepath.EvalSymlinks(path)
	return target
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
	envelopeJSON, _ := json.Marshal(envelope)
	if !strings.Contains(string(envelopeJSON), `"`+wanted+`"`) {
		t.Errorf("%s does not answer %q", envelopeJSON, wanted)
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
			if _, code := codex("ls", "PermissionRequest").response(verdict); code != 0 {
				t.Errorf("%s -> exit %d, wanted 0", verdict, code)
			}
		}
	})
}

func TestStdinToStdout(t *testing.T) {
	t.Run("writes the envelope and the blocking exit code", func(t *testing.T) {
		output, code := outcomeOf(t, inputOf(claude("ls", "PreToolUse")), deny)
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		says(t, envelopeIn(t, output), "decision", "block")
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
		output, code := outcomeOf(t, inputWithTrailer("ls", "  \n\n"), deny)
		if code != 2 {
			t.Errorf("exit %d wrote %q, wanted 2", code, output)
		}
	})

	t.Run("a null payload is ignored", func(t *testing.T) {
		writesNothing(t, strings.NewReader("null"), deny)
	})
}

func leavesUndecided(t *testing.T, payload toolCall) {
	t.Helper()
	if verdict, isAsked := payload.verdict(always(deny)); isAsked {
		t.Errorf("%v -> %s, wanted no opinion", payload, verdict)
	}
}

func responseTo(t *testing.T, payload toolCall, verdict policy.Verdict) (map[string]any, int) {
	t.Helper()
	envelope, code := payload.response(verdict)
	if envelope == nil {
		t.Fatalf("%v -> nothing, wanted an envelope", payload)
	}
	return envelope, code
}

func staysQuiet(t *testing.T, payload toolCall, verdict policy.Verdict, wanted int) {
	t.Helper()
	envelope, code := payload.response(verdict)
	if envelope != nil || code != wanted {
		t.Errorf("%v -> %v, exit %d, wanted nothing and %d", payload, envelope, code, wanted)
	}
}

func section(t *testing.T, envelope map[string]any, key string) map[string]any {
	t.Helper()
	inner, isMap := envelope[key].(map[string]any)
	if !isMap {
		t.Fatalf("%v has no %s", envelope, key)
	}
	return inner
}

func decision(t *testing.T, envelope map[string]any) map[string]any {
	t.Helper()
	return section(t, section(t, envelope, "hookSpecificOutput"), "decision")
}

func says(t *testing.T, envelope map[string]any, key, wanted string) {
	t.Helper()
	if value, _ := envelope[key].(string); value != wanted {
		t.Errorf("%s = %q, wanted %q", key, value, wanted)
	}
}

func inputWithTrailer(command, after string) io.Reader {
	payloadJSON, _ := json.Marshal(claude(command, "PreToolUse"))
	return strings.NewReader(string(payloadJSON) + after)
}

func inputOf(payload toolCall) *bytes.Reader {
	payloadJSON, _ := json.Marshal(payload)
	return bytes.NewReader(payloadJSON)
}

func outcomeOf(t *testing.T, in io.Reader, verdict policy.Verdict) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := Main(in, &out, &bytes.Buffer{}, always(verdict))
	return out.String(), code
}

func envelopeIn(t *testing.T, output string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("%q: %v", output, err)
	}
	return envelope
}

func writesNothing(t *testing.T, in io.Reader, verdict policy.Verdict) {
	t.Helper()
	if output, code := outcomeOf(t, in, verdict); output != "" || code != 0 {
		t.Errorf("wrote %q and exited %d, wanted nothing and 0", output, code)
	}
}
