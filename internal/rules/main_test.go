// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules_test

import (
	"os"
	"testing"

	"github.com/thehale/tooluse-screener/internal/testgit"
)

func TestMain(m *testing.M) {
	testgit.ClearEnvironment()
	os.Exit(m.Run())
}
