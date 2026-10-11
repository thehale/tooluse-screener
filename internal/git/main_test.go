// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package git_test

import (
	"os"
	"testing"

	"github.com/thehale/tooluse-screener/internal/testgit"
)

func TestMain(m *testing.M) {
	testgit.ClearEnvironment()
	os.Exit(m.Run())
}
