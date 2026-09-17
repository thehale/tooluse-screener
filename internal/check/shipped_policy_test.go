// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package check_test

import (
	"testing"

	"github.com/thehale/tooluse-screener/internal/check"
	"github.com/thehale/tooluse-screener/internal/policy"
)

func TestTheShippedPolicy(t *testing.T) {
	t.Run("vouches for the git that reads and commits", func(t *testing.T) {
		ships(t, check.Allow, "git status", "git commit -m hi", "git diff", "git config --get user.name")
	})

	t.Run("refuses the git config that writes", func(t *testing.T) {
		ships(t, check.Deny, "git config --unset user.name", "git config set user.name x")
	})

	t.Run("refuses a publish script, named or run from any path", func(t *testing.T) {
		ships(t, check.Deny, "bin/publish", "./bin/publish", "publish --dry-run", "publish.sh")
		ships(t, check.Ask, "cat docs/publishing.md", "grep -rn publish lib")
	})

	t.Run("refuses to empty a whole tree", func(t *testing.T) {
		ships(t, check.Deny, "rm -rf /", "rm -rf ~")
	})

	t.Run("trusts no directory until one is named", func(t *testing.T) {
		ships(t, check.Ask, "git -C /anywhere status", "git -C ~/Work status")
	})

	t.Run("leaves the rest to a machine, rather than guessing", func(t *testing.T) {
		ships(t, check.Ask, "git push origin main", "git clone a b")
		ships(t, check.Ask, "gh release create v1.2.3", "gh repo create acme/newthing")
		ships(t, check.Ask, "deploy now", "npm install", "ls -la")
	})
}

func ships(t *testing.T, wanted check.Decision, commands ...string) {
	t.Helper()
	for _, command := range commands {
		if verdict := check.Evaluate(command, policy.Default()); verdict.Decision != wanted {
			t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
		}
	}
}
