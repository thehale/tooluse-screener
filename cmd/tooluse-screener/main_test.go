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

const madeUp = "denied: [shutdown]\nallowed: [git status]\n"

func TestCheckingOneCommand(t *testing.T) {
	policy := written(t, madeUp)

	t.Run("an allowed command prints its verdict and exits 0", func(t *testing.T) {
		out, code := ran(t, "--config-file", policy, "git status")
		if code != 0 || !strings.HasPrefix(out, "allow: ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("a denied command exits 1", func(t *testing.T) {
		out, code := ran(t, "--config-file", policy, "shutdown now")
		if code != 1 || !strings.HasPrefix(out, "deny: ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("an unknown command exits 2", func(t *testing.T) {
		out, code := ran(t, "--config-file", policy, "nmap localhost")
		if code != 2 || !strings.HasPrefix(out, "ask: ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("no command at all is a usage error", func(t *testing.T) {
		if _, code := ran(t, "--config-file", policy); code != usageError {
			t.Errorf("exit %d, wanted %d", code, usageError)
		}
	})

	t.Run("a policy that will not read is an error, not a verdict", func(t *testing.T) {
		out, code := ran(t, "--config-file", missing(t), "ls")
		if code != usageError || out != "" {
			t.Errorf("exit %d said %q", code, out)
		}
	})
}

func TestSayingWhatItIs(t *testing.T) {
	t.Run("asking for help is not an error, however it is asked", func(t *testing.T) {
		for _, asking := range [][]string{{"help"}, {"-h"}, {"-help"}, {"--help"}} {
			out, code := ran(t, asking...)
			if code != 0 {
				t.Errorf("%q exited %d, wanted 0", asking, code)
			}
			if !strings.Contains(out, "Usage: tooluse-screener") {
				t.Errorf("%q wrote %q to stdout", asking, out)
			}
		}
	})

	t.Run("help lists every flag it takes", func(t *testing.T) {
		out, _ := ran(t, "help")
		for _, flag := range []string{"-hook", "-version", "-config-file"} {
			if !strings.Contains(out, flag) {
				t.Errorf("help does not mention %s", flag)
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
		out, code := ran(t, "--version")
		if code != 0 || !strings.HasPrefix(out, "tooluse-screener ") {
			t.Errorf("exit %d said %q", code, out)
		}
		if strings.TrimSpace(strings.TrimPrefix(out, "tooluse-screener ")) == "" {
			t.Errorf("named no version: %q", out)
		}
	})
}

func TestAnsweringAHook(t *testing.T) {
	policy := written(t, madeUp)

	t.Run("a denied command is blocked", func(t *testing.T) {
		out, code := answered(t, policy, "shutdown now")
		if code != 2 {
			t.Errorf("exit %d, wanted 2", code)
		}
		if read(t, out)["decision"] != "block" {
			t.Errorf("wrote %q", out)
		}
	})

	t.Run("an allowed command is allowed", func(t *testing.T) {
		out, code := answered(t, policy, "git status")
		if code != 0 {
			t.Errorf("exit %d, wanted 0", code)
		}
		output, _ := read(t, out)["hookSpecificOutput"].(map[string]any)
		if output["permissionDecision"] != "allow" {
			t.Errorf("wrote %q", out)
		}
	})

	t.Run("an unknown command says nothing", func(t *testing.T) {
		if out, code := answered(t, policy, "nmap localhost"); out != "" || code != 0 {
			t.Errorf("wrote %q and exited %d", out, code)
		}
	})

	t.Run("a policy that will not read enforces nothing, and says so", func(t *testing.T) {
		var out, complaints bytes.Buffer
		code := run([]string{"--hook", "--config-file", missing(t)}, sent("shutdown now"), &out, &complaints)
		if out.String() != "" || code != 0 {
			t.Errorf("wrote %q and exited %d, wanted nothing and 0", out.String(), code)
		}
		if !strings.Contains(complaints.String(), "the policy failed") {
			t.Errorf("stderr said %q", complaints.String())
		}
	})
}

func ran(t *testing.T, arguments ...string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := run(arguments, strings.NewReader(""), &out, &bytes.Buffer{})
	return out.String(), code
}

func answered(t *testing.T, policy, command string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := run([]string{"--hook", "--config-file", policy}, sent(command), &out, &bytes.Buffer{})
	return out.String(), code
}

func sent(command string) *bytes.Reader {
	payload := map[string]any{
		"tool_name":  "Bash",
		"tool_input": map[string]any{"command": command},
	}
	written, _ := json.Marshal(payload)
	return bytes.NewReader(written)
}

func read(t *testing.T, out string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("%q: %v", out, err)
	}
	return envelope
}

func written(t *testing.T, configuration string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.yaml")
	if err := os.WriteFile(path, []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func missing(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "missing.yaml")
}
