// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSpecs(t *testing.T) {
	paths, _ := filepath.Glob(filepath.Join("testdata", "*", "*.test"))
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Setenv("HOME", "/home/me")
			policyYAML, checks := readSpec(t, path)
			policy, err := Parse(policyYAML)
			runSpec(t, checks, policy, err)
		})
	}
}

var separator = regexp.MustCompile(`(?m)^-{80}\r?\n`)

func readSpec(t *testing.T, path string) (policyYAML []byte, checks string) {
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

func runSpec(t *testing.T, checks string, policy Policy, err error) {
	t.Helper()
	lines, lineErr := significantLines(checks)
	wanted, isRefusal := refusalWanted(lines)
	switch {
	case lineErr != nil:
		t.Fatal(lineErr)
	case isRefusal:
		checkRefusal(t, err, wanted)
	case err != nil:
		t.Fatalf("Parse: %v", err)
	default:
		runChecks(t, checks, policy)
	}
}

func refusalWanted(lines []line) (string, bool) {
	if len(lines) == 1 {
		return strings.CutPrefix(lines[0].text, "error: ")
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

func runChecks(t *testing.T, checks string, policy Policy) {
	t.Helper()
	spec, err := ParseSpec(checks)
	if err == nil {
		for _, check := range spec.Test(policy) {
			assertPassed(t, check)
		}
	} else {
		t.Fatal(err)
	}
}

func assertPassed(t *testing.T, check Result) {
	t.Helper()
	if check.Outcome == Fail {
		t.Errorf("line %d: %s\ngot  %s\nwant %s", check.Case.Line, check.Case.Prompt, check.Verdict, check.Case.Want)
	}
}
