// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package testutil

type logEntry struct {
	Level  string
	Msg    string
	Fields map[string]interface{}
}

type FakeLogger struct {
	entries []logEntry
}

func (f *FakeLogger) Debug(msg string, keyVals ...interface{}) {
	f.entries = append(f.entries, buildEntry("DEBUG", msg, keyVals...))
}

func (f *FakeLogger) Info(msg string, keyVals ...interface{}) {
	f.entries = append(f.entries, buildEntry("INFO", msg, keyVals...))
}

func (f *FakeLogger) Warn(msg string, keyVals ...interface{}) {
	f.entries = append(f.entries, buildEntry("WARN", msg, keyVals...))
}

func (f *FakeLogger) Error(msg string, keyVals ...interface{}) {
	f.entries = append(f.entries, buildEntry("ERROR", msg, keyVals...))
}

func (f *FakeLogger) Trace(msg string, keyVals ...interface{}) {
	f.entries = append(f.entries, buildEntry("TRACE", msg, keyVals...))
}

func (f *FakeLogger) All() []logEntry {
	copied := make([]logEntry, len(f.entries))
	for i, e := range f.entries {
		fields := make(map[string]interface{}, len(e.Fields))
		for k, v := range e.Fields {
			fields[k] = v
		}
		copied[i] = logEntry{Level: e.Level, Msg: e.Msg, Fields: fields}
	}
	return copied
}

func buildEntry(level, msg string, keyVals ...interface{}) logEntry {
	fields := make(map[string]interface{}, len(keyVals)/2)
	for i := 0; i+1 < len(keyVals); i += 2 {
		key, ok := keyVals[i].(string)
		if !ok {
			continue
		}
		fields[key] = keyVals[i+1]
	}
	return logEntry{Level: level, Msg: msg, Fields: fields}
}
