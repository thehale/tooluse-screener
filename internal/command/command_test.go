// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command_test

import (
	"slices"
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
)

const held = command.Unquotable

func TestSeparators(t *testing.T) {
	t.Run("one command is one command", func(t *testing.T) {
		found(t, "ls -la", "ls -la")
	})

	t.Run("every separator starts another", func(t *testing.T) {
		for _, separator := range []string{";", "&&", "||", "|", "&", "\n"} {
			found(t, "ls "+separator+" cat foo", "ls", "cat foo")
		}
	})

	t.Run("the longer separator wins", func(t *testing.T) {
		found(t, "ls && cat", "ls", "cat")
		found(t, "ls || cat", "ls", "cat")
	})

	t.Run("an ampersand in a redirection separates nothing", func(t *testing.T) {
		found(t, "ls 2>&1", "ls 2>&1")
		found(t, "ls >&2", "ls >&2")
		found(t, "ls &> out", "ls &> out")
	})

	t.Run("whitespace is tidied up", func(t *testing.T) {
		found(t, "  ls   -la  ;   cat foo  ", "ls -la", "cat foo")
	})

	t.Run("nothing is no commands", func(t *testing.T) {
		found(t, "")
		found(t, "   ")
		found(t, "ls ;; cat", "ls", "cat")
	})

	t.Run("a line continuation joins two lines", func(t *testing.T) {
		found(t, "git \\\nstatus", "git status")
	})
}

func TestSubstitutions(t *testing.T) {
	t.Run("a substitution is its own command", func(t *testing.T) {
		found(t, `echo "$(cat foo)"`, `echo "`+held+`"`, "cat foo")
	})

	t.Run("backticks are too", func(t *testing.T) {
		found(t, "echo `cat foo`", "echo "+held, "cat foo")
	})

	t.Run("process substitutions are too", func(t *testing.T) {
		found(t, "diff <(cat a) <(cat b)", "diff "+held+" "+held, "cat a", "cat b")
	})

	t.Run("nested substitutions are found", func(t *testing.T) {
		found(t, `echo "$(cat "$(cat foo)")"`, `echo "`+held+`"`, `cat "`+held+`"`, "cat foo")
	})

	t.Run("what is left behind cannot join its neighbours", func(t *testing.T) {
		found(t, "ls$(cat foo)blk", "ls"+held+"blk", "cat foo")
	})

	t.Run("arithmetic is not a command", func(t *testing.T) {
		found(t, "echo $((1 + 2))", "echo $((1 + 2))")
	})

	t.Run("a parameter expansion is not a command", func(t *testing.T) {
		found(t, "echo ${HOME}", "echo ${HOME}")
	})

	t.Run("an unterminated substitution is read to the end", func(t *testing.T) {
		found(t, "echo $(cat foo", "echo "+held, "cat foo")
		found(t, "echo `cat foo", "echo "+held, "cat foo")
	})

	t.Run("separators inside a substitution still separate", func(t *testing.T) {
		found(t, `echo "$(ls && cat foo)"`, `echo "`+held+`"`, "ls", "cat foo")
	})
}

func found(t *testing.T, text string, wanted ...string) {
	t.Helper()
	if all := command.All(text); !slices.Equal(all, wanted) {
		t.Errorf("All(%q) = %q, wanted %q", text, all, wanted)
	}
}
