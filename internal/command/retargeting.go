// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
	"strings"
)

func IsRetargeting(name string) bool {
	return slices.Contains(retargetingNames, name) || strings.HasPrefix(name, "GIT_CONFIG")
}

var retargetingNames = []string{
	"GIT_DIR", "GIT_WORK_TREE", "GIT_NAMESPACE", "GIT_COMMON_DIR",
	"GH_REPO", "GH_HOST", "GH_CONFIG_DIR",
	"HOME", "XDG_CONFIG_HOME",
}
