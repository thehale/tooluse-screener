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
	if payload, isRead := payloadOn(in); !isRead {
		return 0
	} else {
		return payload.answer(checker, out, complaints)
	}
}

func (call toolCall) answer(checker Checker, out, complaints io.Writer) int {
	if verdict, isAsked := call.safeVerdict(checker, complaints); !isAsked {
		return 0
	} else {
		return call.reply(verdict, out, complaints)
	}
}

func (call toolCall) reply(verdict policy.Verdict, out, complaints io.Writer) int {
	envelope, code := call.response(verdict)
	write(envelope, out)
	if code == blocked {
		_, _ = fmt.Fprintln(complaints, verdict.Reason)
	}
	return code
}

func (call toolCall) safeVerdict(checker Checker, complaints io.Writer) (verdict policy.Verdict, isAsked bool) {
	defer func() {
		if failure := recover(); failure != nil {
			_, _ = fmt.Fprintf(complaints, "tooluse-screener: the policy failed: %v\n", failure)
			verdict, isAsked = failOpen, true
		}
	}()
	return call.verdict(checker)
}

var failOpen = policy.Verdict{Decision: policy.Ask, Reason: "The policy could not answer"}

func write(envelope map[string]any, out io.Writer) {
	if envelope != nil {
		envelopeJSON, _ := json.Marshal(envelope)
		_, _ = out.Write(envelopeJSON)
	}
}

type toolCall map[string]any

func payloadOn(in io.Reader) (toolCall, bool) {
	var payload map[string]any
	decoder := json.NewDecoder(in)
	if err := decoder.Decode(&payload); err != nil {
		return nil, false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, false
	}
	return payload, payload != nil
}

func (call toolCall) text(key string) string {
	value, _ := call[key].(string)
	return value
}
