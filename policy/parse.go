// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "gopkg.in/yaml.v3"

func Parse(contents []byte) (Policy, error) {
	var configuration file
	if err := yaml.Unmarshal(contents, &configuration); err != nil {
		return Policy{}, err
	}
	return configuration.policy()
}
