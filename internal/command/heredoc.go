// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/lists"
	"mvdan.cc/sh/v3/syntax"
)

func withInertHeredocDropped(file *syntax.File) *syntax.File {
	if heredoc, isInert := inertHeredoc(file); isInert {
		stmt := file.Stmts[0]
		stmt.Redirs = slices.DeleteFunc(stmt.Redirs, func(redirect *syntax.Redirect) bool {
			return redirect == heredoc
		})
	}
	return file
}

func inertHeredoc(file *syntax.File) (*syntax.Redirect, bool) {
	if len(file.Stmts) == 1 && isBareCat(file.Stmts[0]) {
		return heredocWrittenToFiles(file.Stmts[0].Redirs)
	} else {
		return nil, false
	}
}

func isBareCat(stmt *syntax.Stmt) bool {
	call, isCall := stmt.Cmd.(*syntax.CallExpr)
	return isCall && !stmt.Negated && !stmt.Background && !stmt.Coprocess &&
		len(call.Assigns) == 0 && len(call.Args) == 1 && call.Args[0].Lit() == "cat"
}

func heredocWrittenToFiles(redirects []*syntax.Redirect) (*syntax.Redirect, bool) {
	heredocs, writes := lists.Partition(redirects, isHeredoc)
	if len(heredocs) == 1 && hasQuotedDelimiter(heredocs[0]) && len(writes) > 0 && lists.Every(writes, isStdoutToFile) {
		return heredocs[0], true
	} else {
		return nil, false
	}
}

func isHeredoc(redirect *syntax.Redirect) bool {
	return redirect.Op == syntax.Hdoc || redirect.Op == syntax.DashHdoc
}

func hasQuotedDelimiter(heredoc *syntax.Redirect) bool {
	return lists.Some(heredoc.Word.Parts, isQuoting)
}

func isQuoting(wordPart syntax.WordPart) bool {
	switch part := wordPart.(type) {
	case *syntax.SglQuoted, *syntax.DblQuoted:
		return true
	case *syntax.Lit:
		return strings.Contains(part.Value, `\`)
	default:
		return false
	}
}

func isStdoutToFile(redirect *syntax.Redirect) bool {
	isStdout := redirect.N == nil || redirect.N.Value == "1"
	return isStdout && slices.Contains(fileWriteOps, redirect.Op) && isOrdinaryFile(redirect.Word.Lit())
}

var fileWriteOps = []syntax.RedirOperator{syntax.RdrOut, syntax.AppOut, syntax.RdrClob}

func isOrdinaryFile(path string) bool {
	return path != "" && !strings.HasPrefix(path, "/dev/") && !strings.HasPrefix(path, "/proc/")
}
