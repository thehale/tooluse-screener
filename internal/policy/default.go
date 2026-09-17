// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import _ "embed"

func Default() Policy {
	return builtIn
}

//go:embed default.yaml
var written []byte

var builtIn = parsedOrPanic(written)

func parsedOrPanic(written []byte) Policy {
	built, err := Parse(written)
	if err != nil {
		panic(err)
	}
	return built
}
