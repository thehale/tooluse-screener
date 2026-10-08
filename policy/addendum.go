// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type addendumBlock struct {
	rules.Addendum
}

func (a *addendumBlock) UnmarshalYAML(node *yaml.Node) error {
	var fields struct {
		Only   *onlyBlock `yaml:"only"`
		Reason string     `yaml:"reason"`
	}
	if err := node.Decode(&fields); err != nil {
		return err
	} else if fields.Only == nil || strings.TrimSpace(fields.Reason) == "" {
		return refusalAt(node.Line, "an addendum is a mapping with `only` and `reason`")
	} else {
		a.Addendum = rules.Addendum{Only: fields.Only.Scope, Reason: fields.Reason}
		return nil
	}
}
