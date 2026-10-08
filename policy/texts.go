// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type texts []string

func (t *texts) UnmarshalYAML(node *yaml.Node) error {
	var values []string
	for _, item := range nodesIn(node) {
		if !isText(item) {
			return refusalAt(item.Line, "a command or expression is written as text, and this is not: %v", item.Value)
		}
		values = append(values, item.Value)
	}
	*t = values
	return nil
}

func nodesIn(node *yaml.Node) []*yaml.Node {
	if node.Kind == yaml.SequenceNode {
		return node.Content
	} else {
		return []*yaml.Node{node}
	}
}

func isText(node *yaml.Node) bool {
	return node.Tag == "!!str" && strings.TrimSpace(node.Value) != ""
}
