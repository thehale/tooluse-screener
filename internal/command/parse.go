// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"sort"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func All(text string) []Command {
	source := withLinesJoined(text)
	file, err := syntax.NewParser().Parse(strings.NewReader(source), "")
	if err == nil {
		return commands(partsOf(source, file), nil)
	} else {
		return Blind(text)
	}
}

func partsOf(source string, file *syntax.File) []part {
	var parts []part
	syntax.Walk(file, func(node syntax.Node) bool {
		stmt, isStmt := node.(*syntax.Stmt)
		if isStmt {
			if text, isCommand := textOf(source, stmt); isCommand {
				parts = append(parts, part{text})
			}
		}
		return true
	})
	return parts
}

func textOf(source string, stmt *syntax.Stmt) (text string, isCommand bool) {
	if isSimple(stmt.Cmd) {
		bounds := spanOf(stmt)
		return withCutsApplied(source, bounds, cutsIn(stmt.Cmd)), true
	} else {
		return "", false
	}
}

func isSimple(cmd syntax.Command) bool {
	switch cmd.(type) {
	case *syntax.CallExpr, *syntax.DeclClause:
		return true
	default:
		return false
	}
}

type span struct {
	start, end int
}

func spanOf(stmt *syntax.Stmt) span {
	start := stmt.Cmd.Pos()
	if stmt.Negated {
		start = stmt.Pos()
	}
	for _, redirect := range stmt.Redirs {
		if start.After(redirect.Pos()) {
			start = redirect.Pos()
		}
	}
	return span{int(start.Offset()), int(widestEnd(stmt).Offset())}
}

func widestEnd(stmt *syntax.Stmt) syntax.Pos {
	end := stmt.Cmd.End()
	for _, redirect := range stmt.Redirs {
		if redirect.End().After(end) {
			end = redirect.End()
		}
	}
	return end
}

type cut struct {
	span
	replacement string
}

func cutsIn(cmd syntax.Command) []cut {
	var cuts []cut
	syntax.Walk(cmd, func(node syntax.Node) bool {
		found, isSubstitution := substitutionSpan(node)
		if isSubstitution {
			cuts = append(cuts, cut{found, Substitution})
		}
		return !isSubstitution
	})
	return cuts
}

func withCutsApplied(source string, bounds span, cuts []cut) string {
	sort.Slice(cuts, func(i, j int) bool { return cuts[i].start < cuts[j].start })
	var text strings.Builder
	cursor := bounds.start
	for _, one := range cuts {
		text.WriteString(source[cursor:one.start])
		text.WriteString(one.replacement)
		cursor = one.end
	}
	text.WriteString(source[cursor:bounds.end])
	return text.String()
}

func substitutionSpan(node syntax.Node) (found span, isSubstitution bool) {
	switch node.(type) {
	case *syntax.CmdSubst, *syntax.ProcSubst:
		return span{int(node.Pos().Offset()), int(node.End().Offset())}, true
	default:
		return span{}, false
	}
}
