// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSpecifications(t *testing.T) {
	paths, _ := filepath.Glob(filepath.Join("testdata", "*", "*.test"))
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Setenv("HOME", "/home/me")
			policyYAML, checks := readSpecification(t, path)
			policy, err := Parse(policyYAML)
			runSpecification(t, checks, policy, err)
		})
	}
}

var separator = regexp.MustCompile(`(?m)^-{80}\r?\n`)

func readSpecification(t *testing.T, path string) (policyYAML []byte, checks string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	replaced := placeholdersReplaced(t, string(contents))
	bounds := separator.FindStringIndex(replaced)
	if bounds == nil {
		t.Fatalf("%s has no 80-dash separator", path)
	}
	return []byte(replaced[:bounds[0]]), replaced[bounds[1]:]
}

func placeholdersReplaced(t *testing.T, text string) string {
	t.Helper()
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReplacer("$ROOT", t.TempDir(), "$HERE", here).Replace(text)
}

func runSpecification(t *testing.T, checks string, policy Policy, err error) {
	t.Helper()
	lines := significantLines(checks)
	wanted, isRefusal := refusalWanted(lines)
	if isRefusal {
		checkRefusal(t, err, wanted)
	} else if err != nil {
		t.Fatalf("Parse: %v", err)
	} else {
		runChecks(t, lines, policy)
	}
}

func significantLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			lines = append(lines, strings.TrimRight(line, "\r"))
		}
	}
	return lines
}

func refusalWanted(lines []string) (string, bool) {
	if len(lines) == 1 {
		return strings.CutPrefix(lines[0], "error: ")
	} else {
		return "", false
	}
}

func checkRefusal(t *testing.T, err error, wanted string) {
	t.Helper()
	if err == nil {
		t.Errorf("Parse was accepted, wanted it refused with %q", wanted)
	} else if got := err.Error(); got != wanted {
		t.Errorf("error\ngot  %s\nwant %s", got, wanted)
	}
}

func runChecks(t *testing.T, lines []string, policy Policy) {
	t.Helper()
	if len(lines) == 0 || len(lines)%2 != 0 {
		t.Fatalf("a specification is a prompt line and its verdict, in pairs: %q", lines)
	} else {
		for index := 0; index < len(lines); index += 2 {
			checkVerdict(t, policy, lines[index], lines[index+1])
		}
	}
}

var editCall = regexp.MustCompile(`^Edit\((.*)\)$`)

func checkVerdict(t *testing.T, policy Policy, line, wanted string) {
	t.Helper()
	cwd, rest, _ := strings.Cut(line, " % ")
	t.Chdir(cwd)
	if got := verdictFor(policy, rest).String(); got != wanted {
		t.Errorf("%s\ngot  %s\nwant %s", rest, got, wanted)
	}
}

func verdictFor(policy Policy, rest string) Verdict {
	if path, isEdit := editTarget(rest); isEdit {
		return policy.CheckPath(path)
	} else {
		return policy.CheckCommand(withRealNewlines(rest))
	}
}

func withRealNewlines(text string) string {
	return strings.ReplaceAll(text, `\n`, "\n")
}

func editTarget(rest string) (string, bool) {
	match := editCall.FindStringSubmatch(rest)
	if match != nil {
		return match[1], true
	} else {
		return "", false
	}
}
