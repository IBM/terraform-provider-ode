// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestStructuredLoggerWrites(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, WithColor(false))
	log.Info("hello", "k", 1)

	out := buf.String()
	if !strings.Contains(out, "INFO") || !strings.Contains(out, "hello") || !strings.Contains(out, "k=1") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestColorCodes(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf)
	log.Error("bad")

	ansi := regexp.MustCompile(`\x1b\[[0-9;]+m`)
	if !ansi.Match(buf.Bytes()) {
		t.Fatalf("expected ANSI color codes in output")
	}
}

func TestWithOptions(t *testing.T) {
	var buf bytes.Buffer
	customFmt := JSON("  ")
	log := New(
		&buf,
		WithColor(false),
		WithLevelColors(map[Level]string{Info: "31"}),
		WithFormatter(customFmt),
	)
	log.Info("multi", "keyVal", 1)

	out := buf.String()
	if !strings.Contains(out, `"keyVal": 1`) {
		t.Fatalf("formatter not applied: %s", out)
	}
}

func TestClampUnknownLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, WithColor(false))
	log.Log(99, "hi")

	if !strings.Contains(buf.String(), "UNKNOWN") {
		t.Fatalf("unexpected level handling: %s", buf.String())
	}
}
