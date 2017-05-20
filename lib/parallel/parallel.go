// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package parallel

import (
	"os/exec"
	"time"
)

// Event is an event that happens during the runner's Run call.
type Event struct {
	Type   EventType              `json:"type,omitempty" yaml:"type,omitempty"`
	Time   time.Time              `json:"time,omitempty" yaml:"time,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty" yaml:"fields,omitempty"`
	Error  string                 `json:"error,omitempty" yaml:"error,omitempty"`
}

// RunnerOption is an option for a new Runner.
type RunnerOption func(*runnerOptions)

// WithFastFail returns a RunnerOption that will return error fun
// Run as soon as one of the commands fails.
func WithFastFail() RunnerOption {
	return func(runnerOptions *runnerOptions) {
		runnerOptions.FastFail = true
	}
}

// WithMaxConcurrentCmds returns a RunnerOption that will make the
// Runner only run maxConcurrentCmds at once, or unlimited if 0.
func WithMaxConcurrentCmds(maxConcurrentCmds int) RunnerOption {
	return func(runnerOptions *runnerOptions) {
		runnerOptions.MaxConcurrentCmds = maxConcurrentCmds
	}
}

// WithEventHandler returns a RunnerOption that will use the
// given EventHandler.
func WithEventHandler(eventHandler func(*Event)) RunnerOption {
	return func(runnerOptions *runnerOptions) {
		runnerOptions.EventHandler = eventHandler
	}
}

// WithClock returns a RunnerOption that will make the Runner
// use the given Clock.
func WithClock(clock func() time.Time) RunnerOption {
	return func(runnerOptions *runnerOptions) {
		runnerOptions.Clock = clock
	}
}

// Runner runs the commands.
type Runner interface {
	// Run the commands.
	//
	// Return error if there was an initialization error, or any of
	// the running commands returned with a non-zero exit code.
	Run(cmds []*exec.Cmd) error
}

// NewRunner returns a new Runner.
func NewRunner(options ...RunnerOption) Runner {
	return newRunner(options...)
}
