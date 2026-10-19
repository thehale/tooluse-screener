// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Rule struct {
	Name       string
	Reason     string
	Expression Expression
	Only       Scope
	Addendums  []Addendum
	Trusted    *TrustedGitDirectories
}

func (r Rule) InDir(workingDir directories.WorkingDir) Rule {
	r.Only = r.Only.InDir(workingDir)
	r.Addendums = lists.Map(r.Addendums, func(addendum Addendum) Addendum { return addendum.InDir(workingDir) })
	r.Trusted = r.trustedInDir(workingDir)
	return r
}

func (r Rule) trustedInDir(workingDir directories.WorkingDir) *TrustedGitDirectories {
	if r.Trusted == nil {
		return nil
	} else {
		trusted := r.Trusted.InDir(workingDir)
		return &trusted
	}
}

func (r Rule) Find(command commands.Command) (Match, bool) {
	if r.Only.isInScope(command) && r.isTrusted(command) {
		return r.Expression.Find(command)
	} else {
		return Match{}, false
	}
}

func (r Rule) isTrusted(command commands.Command) bool {
	return r.Trusted == nil || r.Trusted.hasTargetsOf(command)
}

func (r Rule) ReasonFor(command commands.Command) string {
	reasons := []string{r.Reason}
	for _, addendum := range r.Addendums {
		if addendum.Only.isInScope(command) {
			reasons = append(reasons, addendum.Reason)
		}
	}
	return strings.TrimSpace(strings.Join(reasons, " "))
}

func (r Rule) String() string {
	return nameOr(r.Name, r.Expression.String())
}
