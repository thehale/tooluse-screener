// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command_test

import (
	"slices"
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
)

const substitution = command.Substitution

func TestSeparators(t *testing.T) {
	t.Run("one command is one command", func(t *testing.T) {
		finds(t, "ls -la", "ls -la")
	})

	t.Run("every separator starts another", func(t *testing.T) {
		for _, separator := range []string{";", "&&", "||", "|", "&", "\n"} {
			finds(t, "ls "+separator+" cat foo", "ls", "cat foo")
		}
	})

	t.Run("the longer separator wins", func(t *testing.T) {
		finds(t, "ls && cat", "ls", "cat")
		finds(t, "ls || cat", "ls", "cat")
	})

	t.Run("an ampersand in a redirection separates nothing", func(t *testing.T) {
		finds(t, "ls 2>&1", "ls 2>&1")
		finds(t, "ls >&2", "ls >&2")
		finds(t, "ls &> out", "ls &> out")
	})

	t.Run("whitespace is tidied up", func(t *testing.T) {
		finds(t, "  ls   -la  ;   cat foo  ", "ls -la", "cat foo")
	})

	t.Run("nothing is no commands", func(t *testing.T) {
		finds(t, "")
		finds(t, "   ")
		finds(t, "ls ;; cat", "ls", "cat")
	})

	t.Run("a line continuation joins two lines", func(t *testing.T) {
		finds(t, "git \\\nstatus", "git status")
	})

	t.Run("a quoted separator stays part of its command", func(t *testing.T) {
		finds(t, "grep -E 'a|b' file", "grep -E 'a|b' file")
		finds(t, `grep -E "a|b" file`, `grep -E "a|b" file`)
		finds(t, `grep -E a\|b file`, `grep -E a\|b file`)
		finds(t, "ls 'a&&b'", "ls 'a&&b'")
		finds(t, "ls 'a;b'", "ls 'a;b'")
	})

	t.Run("a real separator after a quoted one still separates", func(t *testing.T) {
		finds(t, "cat 'a|b' | cat", "cat 'a|b'", "cat")
	})

	t.Run("a compound command finds the simple commands inside it", func(t *testing.T) {
		finds(t, "if true; then ls; fi", "true", "ls")
	})

	t.Run("a backgrounded redirection keeps its redirect and loses its ampersand", func(t *testing.T) {
		finds(t, "ls 2>&1 &", "ls 2>&1")
	})

	t.Run("an assignment kept for the rest of the line is still one command", func(t *testing.T) {
		finds(t, "export GIT_DIR=/x", "export GIT_DIR=/x")
	})

	t.Run("an unterminated quote falls back to the quote-blind split", func(t *testing.T) {
		finds(t, "ls 'a;b", "ls 'a", "b")
	})
}

func TestSubstitutions(t *testing.T) {
	t.Run("a substitution is its own command", func(t *testing.T) {
		finds(t, `echo "$(cat foo)"`, `echo "`+substitution+`"`, "cat foo")
	})

	t.Run("backticks are too", func(t *testing.T) {
		finds(t, "echo `cat foo`", "echo "+substitution, "cat foo")
	})

	t.Run("process substitutions are too", func(t *testing.T) {
		finds(t, "diff <(cat a) <(cat b)", "diff "+substitution+" "+substitution, "cat a", "cat b")
	})

	t.Run("nested substitutions are found", func(t *testing.T) {
		finds(t, `echo "$(cat "$(cat foo)")"`, `echo "`+substitution+`"`, `cat "`+substitution+`"`, "cat foo")
	})

	t.Run("what is left behind cannot join its neighbours", func(t *testing.T) {
		finds(t, "ls$(cat foo)blk", "ls"+substitution+"blk", "cat foo")
	})

	t.Run("arithmetic is not a command", func(t *testing.T) {
		finds(t, "echo $((1 + 2))", "echo $((1 + 2))")
	})

	t.Run("a parameter expansion is not a command", func(t *testing.T) {
		finds(t, "echo ${HOME}", "echo ${HOME}")
	})

	t.Run("an unterminated substitution is read to the end", func(t *testing.T) {
		finds(t, "echo $(cat foo", "echo "+substitution, "cat foo")
		finds(t, "echo `cat foo", "echo "+substitution, "cat foo")
	})

	t.Run("separators inside a substitution still separate", func(t *testing.T) {
		finds(t, `echo "$(ls && cat foo)"`, `echo "`+substitution+`"`, "ls", "cat foo")
	})

	t.Run("a substitution opener inside single quotes is not extracted", func(t *testing.T) {
		finds(t, "echo '$(cat foo)'", "echo '$(cat foo)'")
		finds(t, "echo '`cat foo`'", "echo '`cat foo`'")
	})

	t.Run("an escaped opener is a syntax error that falls back to the quote-blind split", func(t *testing.T) {
		finds(t, `echo \$(cat foo)`, `echo \`+substitution, "cat foo")
	})
}

func TestBlind(t *testing.T) {
	wanted := []string{"grep -E 'a", "b' file"}
	if all := texts(command.Blind("grep -E 'a|b' file")); !slices.Equal(all, wanted) {
		t.Errorf("Blind split %q, wanted %q", all, wanted)
	}
}

func finds(t *testing.T, text string, wanted ...string) {
	t.Helper()
	if all := texts(command.All(text)); !slices.Equal(all, wanted) {
		t.Errorf("All(%q) = %q, wanted %q", text, all, wanted)
	}
}

func texts(commands []command.Command) []string {
	var lines []string
	for _, one := range commands {
		lines = append(lines, one.Text)
	}
	return lines
}
