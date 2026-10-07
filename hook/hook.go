// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/thehale/tooluse-screener/policy"
)

func Main(in io.Reader, out, complaints io.Writer, checker Checker) int {
	payload, read := payloadOn(in)
	if !read {
		return 0
	}
	verdict, asked := decided(payload, checker, complaints)
	if !asked {
		return 0
	}
	envelope, code := Respond(payload, verdict)
	wrote(envelope, out)
	if code == blocked {
		_, _ = fmt.Fprintln(complaints, verdict.Reason)
	}
	return code
}

func decided(payload Payload, checker Checker, complaints io.Writer) (verdict policy.Verdict, asked bool) {
	defer func() {
		if failure := recover(); failure != nil {
			_, _ = fmt.Fprintf(complaints, "tooluse-screener: the policy failed: %v\n", failure)
			verdict, asked = failOpen, true
		}
	}()
	return Decide(payload, checker)
}

var failOpen = policy.Verdict{Decision: policy.Ask, Reason: "The policy could not answer"}

func wrote(envelope map[string]any, out io.Writer) {
	if envelope != nil {
		written, _ := json.Marshal(envelope)
		_, _ = fmt.Fprint(out, string(written))
	}
}

type Payload map[string]any

func payloadOn(in io.Reader) (Payload, bool) {
	var payload map[string]any
	reading := json.NewDecoder(in)
	if err := reading.Decode(&payload); err != nil {
		return nil, false
	}
	if _, err := reading.Token(); err != io.EOF {
		return nil, false
	}
	return payload, payload != nil
}

func text(payload Payload, key string) string {
	written, _ := payload[key].(string)
	return written
}
