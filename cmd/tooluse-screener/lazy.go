// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/thehale/tooluse-screener/hook"
	"github.com/thehale/tooluse-screener/policy"
)

type lazyPolicy string

func (c lazyPolicy) InDir(directory string) (hook.Checker, error) {
	if activePolicy, err := policy.Load(string(c)); err != nil {
		return nil, fmt.Errorf("the policy will not read: %w", err)
	} else {
		return activePolicy.InDir(directory), nil
	}
}
