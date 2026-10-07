// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/thehale/tooluse-screener/policy"
)

type lazyPolicy string

func (c lazyPolicy) CheckCommand(line string) policy.Verdict {
	return c.policy().CheckCommand(line)
}

func (c lazyPolicy) CheckPath(path string) policy.Verdict {
	return c.policy().CheckPath(path)
}

func (c lazyPolicy) policy() policy.Policy {
	activePolicy, err := policyAt(string(c))
	if err != nil {
		panic(fmt.Sprintf("the policy will not read: %v", err))
	}
	return activePolicy
}

func policyAt(path string) (policy.Policy, error) {
	if path == "" {
		return policy.Load()
	} else {
		return policy.LoadFile(path)
	}
}
