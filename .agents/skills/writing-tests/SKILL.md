---
name: writing-tests
description: >-
  Write a test for the screener's behaviour. Use when adding or changing a
  `.test` specification, or a Go test.
license: MPL-2.0
---

# Writing tests

A test here teaches. A reader should see one behaviour in it and nothing that
distracts from that behaviour.

## Prefer a specification

Show a verdict the screener reaches in a `.test` specification, which states a
policy and the verdict each command gets under it. Follow the specifications
already there for the format.

Keep a Go test for parsing that a specification cannot express.

## Contrived examples

- Build every example from common unix commands, such as `ls`, `rm`, `echo`,
  `cp` and `mv`. Do not invent a tool, and do not copy a real setup, such as a
  rule for one person's workflow or a branch name.
- Give the policy only the rules the cases reach. Drop an extra allowed entry,
  an extra pattern, or an optional group that no case exercises.
- Keep each command short. Add a pipe or a second command only when the case
  is about one.
- Let the file name state the intent, and write no comments. A comment is
  exceptional, kept for a case that would mislead a careful reader without it.

## Names

Name a file or a test case after the behaviour it shows. Name a command only
when the screener treats that command differently from the others.
