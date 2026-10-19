// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/thehale/tooluse-screener/hook"
	"github.com/thehale/tooluse-screener/policy"
)

type lazyPolicy string

func (c lazyPolicy) InDir(directory string) hook.Checker {
	return c.policy().InDir(directory)
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
