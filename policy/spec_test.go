// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import "testing"

func TestParseSpec(t *testing.T) {
	t.Run("a case parses its directory, prompt and verdict", func(t *testing.T) {
		spec, err := ParseSpec("/some/dir % ls -la\nallow: Every command is allowed: ls\n")
		if err != nil {
			t.Fatal(err)
		}
		if len(spec.Cases) != 1 {
			t.Fatalf("ParseSpec returned %d cases, wanted 1", len(spec.Cases))
		}
		one := spec.Cases[0]
		if one.Line != 1 || one.Directory != "/some/dir" || one.Prompt != "ls -la" {
			t.Errorf("ParseSpec = %+v", one)
		}
		if want := (Verdict{Allow, "Every command is allowed: ls"}); one.Want != want {
			t.Errorf("Want = %+v, wanted %+v", one.Want, want)
		}
	})

	t.Run("a decision other than allow, deny or ask is refused", func(t *testing.T) {
		if _, err := ParseSpec("/some/dir % ls\nmaybe: a reason\n"); err == nil {
			t.Error("ParseSpec was accepted")
		}
	})

	t.Run("a wanted line with no reason is refused", func(t *testing.T) {
		if _, err := ParseSpec("/some/dir % ls\nallow\n"); err == nil {
			t.Error("ParseSpec was accepted")
		}
	})

	t.Run("a policy section is refused", func(t *testing.T) {
		_, err := ParseSpec("--------------------------------------------------------------------------------\n/some/dir % ls\nallow: a reason\n")
		if err == nil || err.Error() != "line 1: a .test file holds no policy" {
			t.Errorf("ParseSpec error = %v", err)
		}
	})

	t.Run("no cases is refused", func(t *testing.T) {
		if _, err := ParseSpec(""); err == nil {
			t.Error("ParseSpec was accepted")
		}
	})
}
