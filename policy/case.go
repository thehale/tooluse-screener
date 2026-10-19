// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"regexp"
	"strings"
)

type Case struct {
	Path      string
	Line      int
	Directory string
	Prompt    string
	Want      Verdict
}

type Outcome string

const (
	Pass Outcome = "pass"
	Fail Outcome = "fail"
)

type Result struct {
	Case    Case
	Verdict Verdict
	Outcome Outcome
}

func (s Specs) Test(p Policy) []Result {
	var results []Result
	for _, spec := range s {
		results = append(results, spec.Test(p)...)
	}
	return results
}

func (s Spec) Test(p Policy) []Result {
	results := make([]Result, len(s.Cases))
	for index, c := range s.Cases {
		results[index] = c.Test(p)
	}
	return results
}

func (c Case) Test(p Policy) Result {
	verdict := verdictFor(p.InDir(c.Directory), c.Prompt)
	return Result{c, verdict, outcomeOf(verdict, c.Want)}
}

func outcomeOf(got, want Verdict) Outcome {
	if got == want {
		return Pass
	} else {
		return Fail
	}
}

var editCall = regexp.MustCompile(`^Edit\((.*)\)$`)

func verdictFor(policy Policy, prompt string) Verdict {
	if path, isEdit := editTarget(prompt); isEdit {
		return policy.CheckPath(path)
	} else {
		return policy.CheckCommand(withRealNewlines(prompt))
	}
}

func editTarget(prompt string) (string, bool) {
	match := editCall.FindStringSubmatch(prompt)
	if match == nil {
		return "", false
	} else {
		return match[1], true
	}
}

func withRealNewlines(text string) string {
	return strings.ReplaceAll(text, `\n`, "\n")
}
