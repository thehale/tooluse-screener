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

var addendumShape = shape{yaml.MappingNode, "a mapping with `only` and `reason`"}

var addendumFields = fields{
	{"only", shape{yaml.MappingNode, "a mapping with `dirs` or `branches`"}},
	{"reason", shape{yaml.ScalarNode, "text"}},
}

func (a *addendumBlock) UnmarshalYAML(node *yaml.Node) error {
	var decoded struct {
		Only   *onlyBlock `yaml:"only"`
		Reason string     `yaml:"reason"`
	}
	if err := addendumShape.check(node, "addendum"); err != nil {
		return err
	} else if err := addendumFields.check(node); err != nil {
		return err
	} else if err := node.Decode(&decoded); err != nil {
		return err
	} else if decoded.Only == nil {
		return refusalAt(node.Line, "addendum missing `only`")
	} else if strings.TrimSpace(decoded.Reason) == "" {
		return refusalAt(node.Line, "addendum missing `reason`")
	} else {
		a.Addendum = rules.Addendum{Only: decoded.Only.Scope, Reason: decoded.Reason}
		return nil
	}
}
