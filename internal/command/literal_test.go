// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command_test

import (
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
)

func TestLiteral(t *testing.T) {
	t.Run("a plain word is itself", func(t *testing.T) {
		literal(t, "/work/app", "/work/app")
		literal(t, "~/work", "~/work")
	})

	t.Run("one pair of quotes around it is taken off", func(t *testing.T) {
		literal(t, `"/work/app"`, "/work/app")
		literal(t, `'/work/app'`, "/work/app")
	})

	t.Run("a word the shell would expand is no literal", func(t *testing.T) {
		for _, word := range []string{`"$REPO"`, "$REPO", "${REPO}", "/work/*", "/work/{a,b}", `/work\ app`, "`pwd`", held} {
			literal(t, word, "")
		}
	})

	t.Run("a quoted tilde is not a home", func(t *testing.T) {
		literal(t, `"~/work"`, "")
	})

	t.Run("quotes that do not pair are no literal", func(t *testing.T) {
		literal(t, `"/work/app`, "")
		literal(t, `"/work/app'`, "")
	})
}

func literal(t *testing.T, word, wanted string) {
	t.Helper()
	if got := command.Literal(word); got != wanted {
		t.Errorf("Literal(%q) = %q, wanted %q", word, got, wanted)
	}
}
