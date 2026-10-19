// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package hook

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/thehale/tooluse-screener/policy"
)

func Main(in io.Reader, out, complaints io.Writer, p Policy) int {
	if payload, isRead := payloadOn(in); !isRead {
		return 0
	} else {
		return payload.answer(p, out, complaints)
	}
}

func (call toolCall) answer(p Policy, out, complaints io.Writer) int {
	if query, isPosed := call.posedQuestion(); !isPosed {
		return 0
	} else {
		return call.reply(call.safeVerdict(query, p, complaints), out, complaints)
	}
}

func (call toolCall) reply(verdict policy.Verdict, out, complaints io.Writer) int {
	envelope, code := call.response(verdict)
	write(envelope, out)
	if code == blockingExitCode {
		_, _ = fmt.Fprintln(complaints, verdict.Reason)
	}
	return code
}

func (call toolCall) safeVerdict(query question, p Policy, complaints io.Writer) (verdict policy.Verdict) {
	defer func() {
		if failure := recover(); failure != nil {
			verdict = failOpen(complaints, failure)
		}
	}()
	if checker, err := p.InDir(call.Directory); err != nil {
		return failOpen(complaints, err)
	} else {
		return query(checker)
	}
}

func failOpen(complaints io.Writer, failure any) policy.Verdict {
	_, _ = fmt.Fprintf(complaints, "tooluse-screener: the policy failed: %v\n", failure)
	return policy.Verdict{Decision: policy.Ask, Reason: "The policy could not answer"}
}

func write(envelope map[string]any, out io.Writer) {
	if envelope != nil {
		envelopeJSON, _ := json.Marshal(envelope)
		_, _ = out.Write(envelopeJSON)
	}
}

type toolCall struct {
	Directory string    `json:"cwd"`
	Event     string    `json:"hook_event_name"`
	Tool      string    `json:"tool_name"`
	Input     toolInput `json:"tool_input"`
}

type toolInput struct {
	Command      string `json:"command"`
	FilePath     string `json:"file_path"`
	NotebookPath string `json:"notebook_path"`
}

func payloadOn(in io.Reader) (toolCall, bool) {
	var payload toolCall
	decoder := json.NewDecoder(in)
	if err := decoder.Decode(&payload); err != nil {
		return toolCall{}, false
	}
	_, err := decoder.Token()
	return payload, err == io.EOF
}
