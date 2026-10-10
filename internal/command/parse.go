// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
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
	var pendingWrites []string
	syntax.Walk(file, func(node syntax.Node) bool {
		stmt, isStmt := node.(*syntax.Stmt)
		if isStmt {
			writes := slices.Concat(pendingWrites, writesOf(source, stmt))
			if text, isCommand := textOf(source, stmt); isCommand {
				parts = append(parts, part{text, writes})
				pendingWrites = nil
			} else {
				pendingWrites = writes
			}
		}
		return true
	})
	return parts
}

func textOf(source string, stmt *syntax.Stmt) (text string, isCommand bool) {
	if isSimple(stmt.Cmd) {
		bounds := spanOf(stmt)
		return withCutsApplied(source, bounds, cutsIn(source, stmt, bounds)), true
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
	end := stmt.Cmd.End()
	for _, redirect := range stmt.Redirs {
		if start.After(redirect.Pos()) {
			start = redirect.Pos()
		}
		if isKept(redirect) && redirect.End().After(end) {
			end = redirect.End()
		}
	}
	return span{int(start.Offset()), int(end.Offset())}
}

func isKept(redirect *syntax.Redirect) bool {
	return isHeredocOrHerestring(redirect.Op)
}

func isHeredocOrHerestring(op syntax.RedirOperator) bool {
	switch op {
	case syntax.Hdoc, syntax.DashHdoc, syntax.WordHdoc:
		return true
	default:
		return false
	}
}

type cut struct {
	span
	replacement string
}

func cutsIn(source string, stmt *syntax.Stmt, bounds span) []cut {
	cuts := redirectCutsIn(stmt, bounds)
	syntax.Walk(stmt.Cmd, func(node syntax.Node) bool {
		found, isSubstitution := substitutionSpan(node)
		if isSubstitution {
			cuts = append(cuts, cut{found, Substitution})
		}
		return !isSubstitution
	})
	return cuts
}

func redirectCutsIn(stmt *syntax.Stmt, bounds span) []cut {
	var cuts []cut
	for _, redirect := range stmt.Redirs {
		if found := spanOfRedirect(redirect); trails(redirect, stmt) && !isKept(redirect) && isWithin(found, bounds) {
			cuts = append(cuts, cut{found, ""})
		}
	}
	return cuts
}

func trails(redirect *syntax.Redirect, stmt *syntax.Stmt) bool {
	return !stmt.Cmd.End().After(redirect.Pos())
}

func spanOfRedirect(redirect *syntax.Redirect) span {
	return span{int(redirect.Pos().Offset()), int(redirect.End().Offset())}
}

func isWithin(inner, outer span) bool {
	return inner.start >= outer.start && inner.end <= outer.end
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

func writesOf(source string, stmt *syntax.Stmt) []string {
	var writes []string
	for _, redirect := range stmt.Redirs {
		if target, isWrite := writeTarget(source, redirect); isWrite {
			writes = append(writes, target)
		}
	}
	return writes
}

func writeTarget(source string, redirect *syntax.Redirect) (target string, isWrite bool) {
	word := source[redirect.Word.Pos().Offset():redirect.Word.End().Offset()]
	if isWriteOp(redirect.Op) && !isDuplicationTarget(word) {
		return Literal(word), true
	} else {
		return "", false
	}
}

func isWriteOp(op syntax.RedirOperator) bool {
	switch op {
	case syntax.RdrOut, syntax.AppOut, syntax.RdrClob, syntax.RdrAll, syntax.AppAll, syntax.RdrInOut, syntax.DplOut:
		return true
	default:
		return false
	}
}

func isDuplicationTarget(word string) bool {
	return word == "-" || isAllDigits(word)
}

func isAllDigits(word string) bool {
	return word != "" && !strings.ContainsFunc(word, isNotADigit)
}

func isNotADigit(letter rune) bool {
	return letter < '0' || letter > '9'
}
