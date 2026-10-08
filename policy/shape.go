// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type shape struct {
	kinds   yaml.Kind
	written string
}

func (s shape) check(node *yaml.Node, subject string) error {
	target := resolved(node)
	if target.Tag == "!!null" || target.Kind&s.kinds != 0 {
		return nil
	} else {
		return refusalAt(target.Line, "%s is %s. Use %s.", subject, described(target), s.written)
	}
}

func resolved(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.AliasNode {
		return node.Alias
	} else {
		return node
	}
}

func described(node *yaml.Node) string {
	switch {
	case node.Kind == yaml.SequenceNode:
		return "a list"
	case node.Kind == yaml.MappingNode:
		return "a mapping"
	case node.Tag == "!!str":
		return fmt.Sprintf("text `%s`", node.Value)
	case node.Tag == "!!int" || node.Tag == "!!float":
		return fmt.Sprintf("number %s", node.Value)
	case node.Tag == "!!null":
		return "empty"
	default:
		return fmt.Sprintf("value %s", node.Value)
	}
}

type field struct {
	key   string
	shape shape
}

type fields []field

func (f fields) check(node *yaml.Node) error {
	mapping := resolved(node)
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if err := f.checkPair(mapping.Content[index], mapping.Content[index+1]); err != nil {
			return err
		}
	}
	return nil
}

func (f fields) checkPair(key, node *yaml.Node) error {
	if expected, isKnown := f.shapeNamed(key.Value); isKnown {
		return expected.check(node, fmt.Sprintf("`%s`", key.Value))
	} else {
		return nil
	}
}

func (f fields) shapeNamed(key string) (shape, bool) {
	for _, one := range f {
		if one.key == key {
			return one.shape, true
		}
	}
	return shape{}, false
}
