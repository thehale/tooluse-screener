// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func Parse(contents []byte) (Policy, error) {
	var configuration file
	if err := yaml.Unmarshal(contents, &configuration); err != nil {
		return Policy{}, refined(err)
	} else {
		return configuration.policy()
	}
}

var linePrefix = regexp.MustCompile(`^line (\d+): `)

func refined(err error) error {
	rest, isFromYAMLV3 := strings.CutPrefix(err.Error(), "yaml: ")
	switch match := linePrefix.FindStringSubmatch(rest); {
	case !isFromYAMLV3:
		return err
	case match != nil:
		line, _ := strconv.Atoi(match[1])
		return refusalAt(line, "invalid YAML: %s", rest[len(match[0]):])
	default:
		return fmt.Errorf("invalid YAML: %s", rest)
	}
}
