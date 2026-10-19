// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"path/filepath"
	"testing"
)

const lsPolicyYAML = `
allowed:
  - ls
`

func TestCaseTest(t *testing.T) {
	allowsLS, err := Parse([]byte(lsPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a verdict matching Want passes", func(t *testing.T) {
		c := Case{Directory: "/path/to/workdir", Prompt: "ls", Want: Verdict{Allow, "Every command is allowed: ls"}}
		check := c.Test(allowsLS)
		if check.Outcome != Pass {
			t.Errorf("Test() = %+v, wanted Pass", check)
		}
	})

	t.Run("a verdict not matching Want fails", func(t *testing.T) {
		c := Case{Directory: "/path/to/workdir", Prompt: "ls", Want: Verdict{Deny, "a reason nobody gave"}}
		check := c.Test(allowsLS)
		if check.Outcome != Fail {
			t.Errorf("Test() = %+v, wanted Fail", check)
		}
	})

	t.Run("a directory that does not exist is judged all the same", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")
		c := Case{Directory: missing, Prompt: "ls", Want: Verdict{Allow, "Every command is allowed: ls"}}
		if check := c.Test(allowsLS); check.Outcome != Pass {
			t.Errorf("Test() = %+v, wanted Pass", check)
		}
	})
}
