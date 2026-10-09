// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "github.com/thehale/tooluse-screener/internal/rules"

type entries []entry

func (e entries) side() (side, error) {
	commandRules, err := e.commandRules()
	if err != nil {
		return side{}, err
	}
	pathRules, err := e.pathRules()
	return side{commandRules: commandRules, pathRules: pathRules}, err
}

func (e entries) pathRules() (rules.Globs, error) {
	var pathRules rules.Globs
	for _, one := range e {
		list, err := one.pathRules()
		if err != nil {
			return nil, err
		}
		pathRules = append(pathRules, list...)
	}
	return pathRules, nil
}

func (e entries) commandRules() (rules.Group, error) {
	var commandRules rules.Group
	for _, one := range e {
		list, err := one.commandRules()
		if err != nil {
			return nil, err
		}
		commandRules = append(commandRules, list...)
	}
	return commandRules, nil
}
