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
		text, err := textOf(item)
		if err != nil {
			return err
		}
		values = append(values, text)
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

func textOf(node *yaml.Node) (string, error) {
	if node.Kind != yaml.ScalarNode {
		return "", refusalAt(node.Line, "value is %s. Use text.", described(node))
	} else if node.Tag == "!!null" || strings.TrimSpace(node.Value) == "" {
		return "", refusalAt(node.Line, "blank value. Fill it in or remove it.")
	} else if node.Tag != "!!str" {
		return "", refusalAt(node.Line, "%s is not text. Quote it: '%s'.", described(node), node.Value)
	} else {
		return node.Value, nil
	}
}
