// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"errors"
	"regexp"
	"strings"
)

type Spec struct {
	Cases []Case
}

type Specs []Spec

func ParseSpec(text string) (Spec, error) {
	lines, err := significantLines(text)
	if err == nil {
		return specFrom(lines)
	} else {
		return Spec{}, err
	}
}

func specFrom(lines []line) (Spec, error) {
	cases, err := casesFrom(lines)
	return Spec{cases}, err
}

type line struct {
	number int
	text   string
}

var separatorLine = regexp.MustCompile(`^-{80}$`)

func significantLines(text string) ([]line, error) {
	var lines []line
	for index, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimRight(raw, "\r")
		significant := strings.TrimSpace(trimmed)
		switch {
		case significant == "" || strings.HasPrefix(significant, "#"):
		case separatorLine.MatchString(significant):
			return nil, refusalAt(index+1, "a .test file holds no policy")
		default:
			lines = append(lines, line{index + 1, trimmed})
		}
	}
	return lines, nil
}

func casesFrom(lines []line) ([]Case, error) {
	switch {
	case len(lines) == 0:
		return nil, errors.New("no cases")
	case len(lines)%2 != 0:
		return nil, refusalAt(lines[len(lines)-1].number, "no verdict after this line")
	default:
		return pairedCases(lines)
	}
}

func pairedCases(lines []line) ([]Case, error) {
	cases := make([]Case, len(lines)/2)
	errs := make([]error, len(lines)/2)
	for index := 0; index < len(lines); index += 2 {
		cases[index/2], errs[index/2] = caseFrom(lines[index], lines[index+1])
	}
	return cases, errors.Join(errs...)
}

func caseFrom(prompt, verdict line) (Case, error) {
	directory, text, _ := strings.Cut(prompt.text, " % ")
	want, err := verdictFrom(verdict)
	return Case{Line: prompt.number, Directory: directory, Prompt: text, Want: want}, err
}

var decisions = map[string]Decision{"allow": Allow, "deny": Deny, "ask": Ask}

func verdictFrom(verdict line) (Verdict, error) {
	word, reason, hasReason := strings.Cut(verdict.text, ": ")
	decision, isDecision := decisions[word]
	if hasReason && isDecision {
		return Verdict{decision, reason}, nil
	} else {
		return Verdict{}, refusalAt(verdict.number, "%q is not a verdict. Write allow/deny/ask: reason.", verdict.text)
	}
}
