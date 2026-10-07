// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheCorpus(t *testing.T) {
	asked := testPolicy(t)
	for _, line := range corpus(t, "corpus.txt") {
		wanted, command := readLine(t, line)
		if verdict := asked.CheckCommand(command); verdict.Decision != wanted {
			t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
		}
	}
}

func TestThePathCorpus(t *testing.T) {
	t.Setenv("HOME", "/home/me")
	asked := testPolicy(t)
	for _, line := range corpus(t, "paths.txt") {
		wanted, path := readLine(t, line)
		if answered := decidedOn(path, asked); answered != wanted {
			t.Errorf("%q -> %s, wanted %s", path, answered, wanted)
		}
	}
}

func testPolicy(t *testing.T) Policy {
	t.Helper()
	asked, err := LoadFile(filepath.Join("testdata", "policy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return asked
}

func decidedOn(path string, asked Policy) Decision {
	if verdict, answered := asked.CheckPath(path); answered {
		return verdict.Decision
	} else {
		return unanswered
	}
}

const unanswered Decision = "none"

func corpus(t *testing.T, named string) []string {
	t.Helper()
	written, err := os.ReadFile(filepath.Join("testdata", named))
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	for _, line := range strings.Split(string(written), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			found = append(found, trimmed)
		}
	}
	return found
}

func readLine(t *testing.T, line string) (Decision, string) {
	t.Helper()
	wanted, command, spelled := strings.Cut(line, " ")
	if !spelled {
		t.Fatalf("a corpus line is a verdict and a command: %q", line)
	}
	return Decision(wanted), strings.TrimSpace(command)
}
