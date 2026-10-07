// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import _ "embed"

func Default() Policy {
	return builtIn
}

//go:embed default.yaml
var defaultYAML []byte

var builtIn = mustParse(defaultYAML)

func mustParse(contents []byte) Policy {
	policy, err := Parse(contents)
	if err != nil {
		panic(err)
	}
	return policy
}
