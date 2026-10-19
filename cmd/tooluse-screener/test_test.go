// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const lsPolicyYAML = `
allowed:
  - ls
`

func TestRunningSpecs(t *testing.T) {
	policy := policyFile(t, lsPolicyYAML)

	t.Run("a passing file exits 0 and prints ok lines", func(t *testing.T) {
		spec := specFile(t, "/ % ls\nallow: Every command is allowed: ls\n")
		out, code := outcomeOf(t, "--config-file", policy, "test", spec)
		passed := strings.Contains(out, "ok   "+spec+":1  / % ls") &&
			strings.Contains(out, "1 case. 1 passed. 0 failed")
		if code != 0 || !passed {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("a failing case exits 1 and prints FAIL with got and want", func(t *testing.T) {
		spec := specFile(t, "/ % ls\nallow: a reason nobody gave\n")
		out, code := outcomeOf(t, "--config-file", policy, "test", spec)
		failed := strings.Contains(out, "FAIL "+spec+":1  / % ls") &&
			strings.Contains(out, "got  allow: Every command is allowed: ls") &&
			strings.Contains(out, "want allow: a reason nobody gave") &&
			strings.Contains(out, "1 case. 0 passed. 1 failed")
		if code != 1 || !failed {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("no files is a usage error", func(t *testing.T) {
		out, code := outcomeOf(t, "--config-file", policy, "test")
		if code != usageError || out != "" {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("a file carrying a policy section exits 64 and runs nothing", func(t *testing.T) {
		spec := specFile(t, strings.Repeat("-", 80)+"\n/ % ls\nallow: Every command is allowed: ls\n")
		out, complaints, code := runWithComplaints(t, "--config-file", policy, "test", spec)
		named := strings.Contains(complaints, spec) && strings.Contains(complaints, "line 1")
		if code != usageError || out != "" || !named {
			t.Errorf("exit %d wrote %q to stdout and %q to stderr", code, out, complaints)
		}
	})

	t.Run("a policy that will not load exits 64", func(t *testing.T) {
		spec := specFile(t, "/ % ls\nallow: Every command is allowed: ls\n")
		out, code := outcomeOf(t, "--config-file", missingFile(t), "test", spec)
		if code != usageError || out != "" {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("the policy comes from the environment variable", func(t *testing.T) {
		spec := specFile(t, "/ % ls\nallow: Every command is allowed: ls\n")
		t.Setenv("TOOLUSE_SCREENER_POLICY_FILE", policy)
		out, code := outcomeOf(t, "test", spec)
		if code != 0 || !strings.Contains(out, "ok   ") {
			t.Errorf("exit %d said %q", code, out)
		}
	})
}

func TestRunningSpecsInADirectory(t *testing.T) {
	policy := policyFile(t, lsPolicyYAML)
	passing := "/ % ls\nallow: Every command is allowed: ls\n"

	t.Run("a directory with no .test files is a usage error", func(t *testing.T) {
		root := t.TempDir()
		out, code := outcomeOf(t, "--config-file", policy, "test", root)
		if code != usageError || out != "" {
			t.Errorf("exit %d said %q", code, out)
		}
	})

	t.Run("every directory with no .test files is named", func(t *testing.T) {
		first, second := t.TempDir(), t.TempDir()
		_, complaints, code := runWithComplaints(t, "--config-file", policy, "test", first, second)
		if code != usageError || !strings.Contains(complaints, first) || !strings.Contains(complaints, second) {
			t.Errorf("exit %d complained %q", code, complaints)
		}
	})

	t.Run("a file and a directory together both run", func(t *testing.T) {
		root := t.TempDir()
		writeSpecAt(t, filepath.Join(root, "in-dir.test"), passing)
		file := specFile(t, passing)
		out, code := outcomeOf(t, "--config-file", policy, "test", file, root)
		named := strings.Contains(out, file) && strings.Contains(out, filepath.Join(root, "in-dir.test"))
		if code != 0 || !named {
			t.Errorf("exit %d said %q", code, out)
		}
	})
}

func writeSpecAt(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func specFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.test")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runWithComplaints(t *testing.T, arguments ...string) (out, complaints string, code int) {
	t.Helper()
	var outBuffer, complaintsBuffer bytes.Buffer
	code = run(arguments, strings.NewReader(""), &outBuffer, &complaintsBuffer)
	return outBuffer.String(), complaintsBuffer.String(), code
}
