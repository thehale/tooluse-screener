// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const examplePolicyYAML = `
denied:
  - shutdown
allowed:
  - git status
`

func TestCheckingOneCommand(t *testing.T) {
	policy := policyFile(t, examplePolicyYAML)

	t.Run("an allowed command prints its verdict and exits 0", func(t *testing.T) {
		out, code := outcomeOf(t, "--config-file", policy, "git status")
		if code != 0 || !strings.HasPrefix(out, "allow: ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("a denied command exits 1", func(t *testing.T) {
		out, code := outcomeOf(t, "--config-file", policy, "shutdown now")
		if code != 1 || !strings.HasPrefix(out, "deny: ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("an unknown command exits 2", func(t *testing.T) {
		out, code := outcomeOf(t, "--config-file", policy, "nmap localhost")
		if code != 2 || !strings.HasPrefix(out, "ask: ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("no command at all is a usage error", func(t *testing.T) {
		if _, code := outcomeOf(t, "--config-file", policy); code != usageError {
			t.Errorf("exit %d, wanted %d", code, usageError)
		}
	})

	t.Run("a policy that will not read is an error, not a verdict", func(t *testing.T) {
		out, code := outcomeOf(t, "--config-file", missingFile(t), "ls")
		if code != usageError || out != "" {
			t.Errorf("exit %d said %q", code, out)
		}
	})
}

func TestSayingWhatItIs(t *testing.T) {
	t.Run("asking for help is not an error, however it is asked", func(t *testing.T) {
		for _, asking := range [][]string{{"help"}, {"-h"}, {"--help"}} {
			out, code := outcomeOf(t, asking...)
			if code != 0 {
				t.Errorf("%q exited %d, wanted 0", asking, code)
			}
			if !strings.Contains(out, "Usage: tooluse-screener") {
				t.Errorf("%q wrote %q to stdout", asking, out)
			}
		}
	})

	t.Run("help lists every flag it takes", func(t *testing.T) {
		out, _ := outcomeOf(t, "help")
		for _, flag := range []string{"--hook", "--version", "--config-file"} {
			if !strings.Contains(out, flag) {
				t.Errorf("help does not mention %s", flag)
			}
		}
	})

	t.Run("one dash is not how a flag is spelled", func(t *testing.T) {
		for _, badFlag := range []string{"-hook", "-version", "-config-file", "-help"} {
			var out, complaints bytes.Buffer
			code := run([]string{badFlag}, strings.NewReader(""), &out, &complaints)
			if code != usageError {
				t.Errorf("%s exited %d, wanted %d", badFlag, code, usageError)
			}
			if !strings.Contains(complaints.String(), "Write -"+badFlag) {
				t.Errorf("%s was not told how to spell it: %q", badFlag, complaints.String())
			}
		}
	})

	t.Run("a usage error complains on stderr rather than stdout", func(t *testing.T) {
		var out, complaints bytes.Buffer
		code := run([]string{"--nonsense"}, strings.NewReader(""), &out, &complaints)
		if code != usageError || out.String() != "" {
			t.Errorf("exit %d wrote %q to stdout", code, out.String())
		}
		if !strings.Contains(complaints.String(), "Usage: tooluse-screener") {
			t.Errorf("stderr said %q", complaints.String())
		}
	})

	t.Run("it says which build it is", func(t *testing.T) {
		out, code := outcomeOf(t, "--version")
		if code != 0 || !strings.HasPrefix(out, "tooluse-screener ") {
			t.Errorf("exit %d said %q", code, out)
		}
		if strings.TrimSpace(strings.TrimPrefix(out, "tooluse-screener ")) == "" {
			t.Errorf("named no version: %q", out)
		}
	})
}

func TestAnsweringAHook(t *testing.T) {
	policy := policyFile(t, examplePolicyYAML)

	t.Run("a denied command is blocked", func(t *testing.T) {
		out, code := hookOutcome(t, policy, "shutdown now")
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		if envelopeIn(t, out)["decision"] != "block" {
			t.Errorf("wrote %q", out)
		}
	})

	t.Run("an allowed command is allowed", func(t *testing.T) {
		out, code := hookOutcome(t, policy, "git status")
		if code != 0 {
			t.Errorf("exit %d, wanted 0", code)
		}
		output, _ := envelopeIn(t, out)["hookSpecificOutput"].(map[string]any)
		if output["permissionDecision"] != "allow" {
			t.Errorf("wrote %q", out)
		}
	})

	t.Run("an unknown command says nothing", func(t *testing.T) {
		if out, code := hookOutcome(t, policy, "nmap localhost"); out != "" || code != 0 {
			t.Errorf("wrote %q and exited %d", out, code)
		}
	})

	t.Run("a policy that will not read enforces nothing, and says so", func(t *testing.T) {
		var out, complaints bytes.Buffer
		code := run([]string{"--hook", "--config-file", missingFile(t)}, payloadFor("shutdown now"), &out, &complaints)
		if out.String() != "" || code != 0 {
			t.Errorf("wrote %q and exited %d, wanted nothing and 0", out.String(), code)
		}
		if !strings.Contains(complaints.String(), "the policy failed") {
			t.Errorf("stderr said %q", complaints.String())
		}
	})
}

func TestAnsweringAFileWrite(t *testing.T) {
	policy := policyFile(t, "denied:\n  - name: Write a secret\n    reason: Ask first.\n    paths: /secrets/**\n")

	t.Run("a write to a denied path is blocked with its reason", func(t *testing.T) {
		out, code := hookOutcomeForWrite(t, policy, "/secrets/key")
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		if reason := envelopeIn(t, out)["reason"]; reason != "Path matches a denied rule: Write a secret. Ask first." {
			t.Errorf("gave the reason %q", reason)
		}
	})

	t.Run("a write anywhere else says nothing", func(t *testing.T) {
		if out, code := hookOutcomeForWrite(t, policy, "/work/notes.md"); out != "" || code != 0 {
			t.Errorf("wrote %q and exited %d", out, code)
		}
	})
}

func outcomeOf(t *testing.T, arguments ...string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := run(arguments, strings.NewReader(""), &out, &bytes.Buffer{})
	return out.String(), code
}

func hookOutcome(t *testing.T, policy, command string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := run([]string{"--hook", "--config-file", policy}, payloadFor(command), &out, &bytes.Buffer{})
	return out.String(), code
}

func hookOutcomeForWrite(t *testing.T, policy, path string) (string, int) {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{"tool_name": "Write", "tool_input": map[string]any{"file_path": path}})
	var out bytes.Buffer
	code := run([]string{"--hook", "--config-file", policy}, bytes.NewReader(payload), &out, &bytes.Buffer{})
	return out.String(), code
}

func payloadFor(command string) *bytes.Reader {
	payload := map[string]any{
		"tool_name":  "Bash",
		"tool_input": map[string]any{"command": command},
	}
	payloadJSON, _ := json.Marshal(payload)
	return bytes.NewReader(payloadJSON)
}

func envelopeIn(t *testing.T, out string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("%q: %v", out, err)
	}
	return envelope
}

func policyFile(t *testing.T, configuration string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.yaml")
	if err := os.WriteFile(path, []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func missingFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "missing.yaml")
}
