// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"bytes"
	"encoding/json"
	"errors"
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

func (a always) InDir(string) (Checker, error) {
	return a, nil
}

func (a always) CheckCommand(string) policy.Verdict {
	return policy.Verdict(a)
}

func (a always) CheckPath(string) policy.Verdict {
	return policy.Verdict(a)
}

type panics struct{}

func (panics) InDir(string) (Checker, error) {
	panic("something went sideways")
}

func (panics) CheckCommand(string) policy.Verdict {
	panic("something went sideways")
}

func (panics) CheckPath(string) policy.Verdict {
	panic("something went sideways")
}

type unreadable struct{}

func (unreadable) InDir(string) (Checker, error) {
	return nil, errors.New("line 3 is not a rule")
}

type echoing struct{}

func (echoing) CheckCommand(line string) policy.Verdict {
	return policy.Verdict{Decision: policy.Deny, Reason: line}
}

func (echoing) CheckPath(path string) policy.Verdict {
	return policy.Verdict{Decision: policy.Deny, Reason: path}
}

type wherever struct{}

func (wherever) InDir(directory string) (Checker, error) {
	return always{Decision: policy.Deny, Reason: directory}, nil
}

type payload map[string]any

func writeCall(t *testing.T, tool, key, path string) payload {
	return payload{
		"tool_name":       tool,
		"hook_event_name": "PreToolUse",
		"cwd":             t.TempDir(),
		"tool_input":      map[string]any{key: path, "content": "hi"},
	}
}

func claude(command, event string) payload {
	return payload{
		"tool_name":       "Bash",
		"hook_event_name": event,
		"tool_input":      map[string]any{"command": command},
	}
}

func codex(command, event string) payload {
	codexPayload := claude(command, event)
	codexPayload["model"] = "gpt-5.6-sol"
	codexPayload["tool_use_id"] = "exec-0e2a"
	return codexPayload
}

func TestNotOurBusiness(t *testing.T) {
	t.Run("a tool that is not bash", func(t *testing.T) {
		leavesUndecided(t, payload{"tool_name": "Read", "tool_input": map[string]any{"file_path": "/etc/passwd"}})
	})

	t.Run("a bash call with no command", func(t *testing.T) {
		leavesUndecided(t, payload{"tool_name": "Bash", "tool_input": map[string]any{}})
	})

	t.Run("a bash call with an empty command", func(t *testing.T) {
		leavesUndecided(t, claude("", "PreToolUse"))
	})

	t.Run("a payload with no tool at all", func(t *testing.T) {
		leavesUndecided(t, payload{})
	})

	t.Run("a tool input that is not an object", func(t *testing.T) {
		for _, input := range []any{nil, []any{}, "ls", 7} {
			leavesUndecided(t, payload{"tool_name": "Bash", "tool_input": input})
		}
	})

	t.Run("a command that is not a string", func(t *testing.T) {
		for _, input := range []any{nil, []any{}, map[string]any{}, 7} {
			leavesUndecided(t, payload{"tool_name": "Bash", "tool_input": map[string]any{"command": input}})
		}
	})

	t.Run("a cwd that is not a string", func(t *testing.T) {
		for _, cwd := range []any{[]any{}, map[string]any{}, 7} {
			call := claude("ls", "PreToolUse")
			call["cwd"] = cwd
			leavesUndecided(t, call)
		}
	})

	t.Run("an event name that is not a string", func(t *testing.T) {
		for _, event := range []any{[]any{}, map[string]any{}, 7} {
			call := claude("ls", "PreToolUse")
			call["hook_event_name"] = event
			leavesUndecided(t, call)
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
		asks(t, call, filepath.Join(call["cwd"].(string), "notes.md"))
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

func asks(t *testing.T, sent payload, wanted string) {
	t.Helper()
	if query, isPosed := callOf(t, sent).posedQuestion(); !isPosed {
		t.Errorf("%v asked nothing, wanted %q", sent, wanted)
	} else if reason := query(echoing{}).Reason; reason != wanted {
		t.Errorf("%v asked about %q, wanted %q", sent, reason, wanted)
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

	t.Run("to read is treated as ask, and its error said on stderr", func(t *testing.T) {
		var out, complaints bytes.Buffer
		code := Main(inputOf(claude("ls", "PreToolUse")), &out, &complaints, unreadable{})
		if code != 0 || out.String() != "" {
			t.Errorf("exit %d wrote %q, wanted a quiet 0", code, out.String())
		}
		if !strings.Contains(complaints.String(), "line 3 is not a rule") {
			t.Errorf("stderr said %q", complaints.String())
		}
	})
}

func TestJudgingFromTheSessionDirectory(t *testing.T) {
	t.Run("a bash call is judged in the payload's cwd", func(t *testing.T) {
		call := claude("ls", "PreToolUse")
		call["cwd"] = "/path/to/workdir"
		judgedIn(t, call, "/path/to/workdir")
	})

	t.Run("no cwd leaves the policy in the process's directory", func(t *testing.T) {
		judgedIn(t, claude("ls", "PreToolUse"), "")
	})

	t.Run("the process stays where it was", func(t *testing.T) {
		before, _ := os.Getwd()
		call := claude("ls", "PreToolUse")
		call["cwd"] = t.TempDir()
		judgedIn(t, call, call["cwd"].(string))
		if after, _ := os.Getwd(); after != before {
			t.Errorf("working directory moved from %s to %s", before, after)
		}
	})
}

func judgedIn(t *testing.T, sent payload, directory string) {
	t.Helper()
	var out, complaints bytes.Buffer
	Main(inputOf(sent), &out, &complaints, wherever{})
	output := section(t, envelopeIn(t, out.String()), "hookSpecificOutput")
	if reason, _ := output["permissionDecisionReason"].(string); reason != directory {
		t.Errorf("judged in %q, wanted %q", reason, directory)
	}
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
			call := claude("git status", wanted.event)
			if wanted.envelope == "" {
				staysQuiet(t, call, wanted.verdict, wanted.code)
			} else {
				envelope, code := responseTo(t, call, wanted.verdict)
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
		envelope, _ := responseTo(t, payload{"tool_name": "Bash", "tool_input": map[string]any{"command": "ls"}}, deny)
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
			if _, code := callOf(t, codex("ls", "PermissionRequest")).response(verdict); code != 0 {
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

func callOf(t *testing.T, sent payload) toolCall {
	t.Helper()
	call, isRead := payloadOn(inputOf(sent))
	if !isRead {
		t.Fatalf("%v did not read as a tool call", sent)
	}
	return call
}

func leavesUndecided(t *testing.T, sent payload) {
	t.Helper()
	for _, verdict := range []policy.Verdict{allow, deny} {
		writesNothing(t, inputOf(sent), verdict)
	}
}

func responseTo(t *testing.T, sent payload, verdict policy.Verdict) (map[string]any, int) {
	t.Helper()
	envelope, code := callOf(t, sent).response(verdict)
	if envelope == nil {
		t.Fatalf("%v -> nothing, wanted an envelope", sent)
	}
	return envelope, code
}

func staysQuiet(t *testing.T, sent payload, verdict policy.Verdict, wanted int) {
	t.Helper()
	envelope, code := callOf(t, sent).response(verdict)
	if envelope != nil || code != wanted {
		t.Errorf("%v -> %v, exit %d, wanted nothing and %d", sent, envelope, code, wanted)
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

func inputOf(sent payload) *bytes.Reader {
	payloadJSON, _ := json.Marshal(sent)
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
