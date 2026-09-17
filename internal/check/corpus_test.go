// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package check_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thehale/tooluse-screener/internal/check"
	"github.com/thehale/tooluse-screener/internal/policy"
)

func TestTheCorpus(t *testing.T) {
	asked, err := policy.Read(filepath.Join("testdata", "policy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range corpus(t) {
		wanted, command := readLine(t, line)
		if verdict := check.Evaluate(command, asked); verdict.Decision != wanted {
			t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
		}
	}
}

func corpus(t *testing.T) []string {
	t.Helper()
	written, err := os.ReadFile(filepath.Join("testdata", "corpus.txt"))
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

func readLine(t *testing.T, line string) (check.Decision, string) {
	t.Helper()
	wanted, command, spelled := strings.Cut(line, " ")
	if !spelled {
		t.Fatalf("a corpus line is a verdict and a command: %q", line)
	}
	return check.Decision(wanted), strings.TrimSpace(command)
}
