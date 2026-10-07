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
	corpusPolicy := testPolicy(t)
	for _, line := range corpus(t, "corpus.txt") {
		wanted, command := readLine(t, line)
		if verdict := corpusPolicy.CheckCommand(command); verdict.Decision != wanted {
			t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
		}
	}
}

func TestThePathCorpus(t *testing.T) {
	t.Setenv("HOME", "/home/me")
	corpusPolicy := testPolicy(t)
	for _, line := range corpus(t, "paths.txt") {
		wanted, path := readLine(t, line)
		if verdict := corpusPolicy.CheckPath(path); verdict.Decision != wanted {
			t.Errorf("%q -> %s, wanted %s", path, verdict, wanted)
		}
	}
}

func testPolicy(t *testing.T) Policy {
	t.Helper()
	corpusPolicy, err := LoadFile(filepath.Join("testdata", "policy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return corpusPolicy
}

func corpus(t *testing.T, fileName string) []string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(string(contents), "\n") {
		if entry := strings.TrimSpace(line); entry != "" && !strings.HasPrefix(entry, "#") {
			lines = append(lines, entry)
		}
	}
	return lines
}

func readLine(t *testing.T, line string) (Decision, string) {
	t.Helper()
	wanted, command, hasCommand := strings.Cut(line, " ")
	if !hasCommand {
		t.Fatalf("a corpus line is a verdict and a command: %q", line)
	}
	return Decision(wanted), strings.TrimSpace(command)
}
