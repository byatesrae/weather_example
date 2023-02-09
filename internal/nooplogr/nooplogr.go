// Package nooplogr allows the instantiation of a no-op implementation of logr.
package nooplogr

import "github.com/go-logr/logr"

type noopSink struct{}

func (s noopSink) Init(info logr.RuntimeInfo)                                {}
func (s noopSink) Enabled(level int) bool                                    { return false }
func (s noopSink) Info(level int, msg string, keysAndValues ...interface{})  {}
func (s noopSink) Error(err error, msg string, keysAndValues ...interface{}) {}
func (s noopSink) WithValues(keysAndValues ...interface{}) logr.LogSink      { return s }
func (s noopSink) WithName(name string) logr.LogSink                         { return s }

// New returns a noop logr (with an underlying noop sink).
func New() logr.Logger {
	return logr.New(noopSink{})
}
