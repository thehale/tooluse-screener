// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func All(text string) []Command {
	source := withLinesJoined(text)
	file, err := syntax.NewParser().Parse(strings.NewReader(source), "")
	if err == nil {
		return commandsFrom(partsOf(source, file))
	} else {
		return Blind(text)
	}
}

func partsOf(source string, file *syntax.File) []string {
	var parts []string
	syntax.Walk(file, func(node syntax.Node) bool {
		if part, isCommand := textOf(source, node); isCommand {
			parts = append(parts, part)
		}
		return true
	})
	return parts
}

func textOf(source string, node syntax.Node) (text string, isCommand bool) {
	stmt, isStmt := node.(*syntax.Stmt)
	if isStmt && isSimple(stmt.Cmd) {
		return withSubstitutionsHidden(source, spanOf(stmt), stmt.Cmd), true
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

func withSubstitutionsHidden(source string, bounds span, cmd syntax.Command) string {
	var text strings.Builder
	cursor := bounds.start
	syntax.Walk(cmd, func(node syntax.Node) bool {
		found, isSubstitution := substitutionSpan(node)
		if isSubstitution {
			text.WriteString(source[cursor:found.start])
			text.WriteString(Unquotable)
			cursor = found.end
		}
		return !isSubstitution
	})
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
