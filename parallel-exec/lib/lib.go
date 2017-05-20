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

package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// EventTypeStarted says that the runner started.
	EventTypeStarted EventType = iota
	// EventTypeCmdStarted says that a command started.
	EventTypeCmdStarted
	// EventTypeCmdFinished says that a command finished.
	EventTypeCmdFinished
	// EventTypeFinished says that the runner finished.
	EventTypeFinished

	_defaultFastFail = false
)

var (
	_defaultMaxConcurrentCmds = runtime.NumCPU()
	_defaultEventHandler      = logEvent
	_defaultClock             = time.Now

	errCmdFailed   = errors.New("command failed")
	errInterrupted = errors.New("interrupted")
)

// EventType is an event type during the runner's run call.
type EventType int

// String returns a string representation of the EventType.
func (e EventType) String() string {
	switch e {
	case EventTypeStarted:
		return "started"
	case EventTypeCmdStarted:
		return "cmd_started"
	case EventTypeCmdFinished:
		return "cmd_finished"
	case EventTypeFinished:
		return "finished"
	default:
		return strconv.Itoa(int(e))
	}
}

// UnmarshalText parses an EventType from it's string representation.
func (e *EventType) UnmarshalText(text []byte) error {
	textString := strings.ToLower(string(text))
	switch textString {
	case "started":
		*e = EventTypeStarted
	case "cmd_started":
		*e = EventTypeCmdStarted
	case "cmd_finished":
		*e = EventTypeCmdFinished
	case "finished":
		*e = EventTypeFinished
	default:
		return fmt.Errorf("unknown EventType: %s", textString)
	}
	return nil
}

// Event is an event that happens during the runner's run call.
type Event interface {
	Type() EventType
	Time() time.Time
	Fields() map[string]string
}

// EventHandler handles events.
type EventHandler func(Event)

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
func WithEventHandler(eventHandler EventHandler) RunnerOption {
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

type event struct {
	E EventType
	T time.Time
	F map[string]string
}

func newEvent(e EventType, t time.Time, f map[string]string) *event {
	return &event{e, t, f}
}

func newStartedEvent(t time.Time) *event {
	return newEvent(EventTypeStarted, t, nil)
}

func newCmdStartedEvent(t time.Time, cmd *exec.Cmd) *event {
	return newEvent(EventTypeCmdStarted, t, map[string]string{
		"cmd": cmdString(cmd),
	})
}

func newCmdFinishedEvent(t time.Time, cmd *exec.Cmd, startTime time.Time, err error) *event {
	f := map[string]string{
		"cmd":      cmdString(cmd),
		"duration": t.Sub(startTime).String(),
	}
	if err != nil {
		f["err"] = err.Error()
	}
	return newEvent(EventTypeCmdFinished, t, f)
}

func newFinishedEvent(t time.Time, startTime time.Time, err error) *event {
	f := map[string]string{
		"duration": t.Sub(startTime).String(),
	}
	if err != nil {
		f["err"] = err.Error()
	}
	return newEvent(EventTypeFinished, t, f)
}

func (e *event) Type() EventType {
	return e.E
}

func (e *event) Time() time.Time {
	return e.T
}

func (e *event) Fields() map[string]string {
	fields := make(map[string]string, 0)
	for key, value := range e.F {
		fields[key] = value
	}
	return fields
}

func (e *event) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":   e.E.String(),
		"time":   e.T,
		"fields": e.F,
	})
}

type runnerOptions struct {
	FastFail          bool
	MaxConcurrentCmds int
	EventHandler      EventHandler
	Clock             func() time.Time
}

func newRunnerOptions() *runnerOptions {
	return &runnerOptions{
		_defaultFastFail,
		_defaultMaxConcurrentCmds,
		_defaultEventHandler,
		_defaultClock,
	}
}

type runner struct {
	options *runnerOptions
}

func newRunner(options ...RunnerOption) *runner {
	runnerOptions := newRunnerOptions()
	for _, option := range options {
		option(runnerOptions)
	}
	return &runner{runnerOptions}
}

func (r *runner) Run(cmds []*exec.Cmd) error {
	// do not want to acquire lock in the signal handler
	// do there is a race condition where err could be set to
	// errCmdFailed or not set at all even after an interrupt happens
	var err error
	doneC := make(chan struct{})
	cmdControllers := make([]*cmdController, len(cmds))
	for i, cmd := range cmds {
		cmdControllers[i] = newCmdController(cmd, r.options.EventHandler, r.options.Clock)
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt)
	go func() {
		for _ = range signalC {
			err = errInterrupted
			doneC <- struct{}{}
			return
		}
	}()

	var waitGroup sync.WaitGroup
	semaphore := newSemaphore(r.options.MaxConcurrentCmds)

	startTime := r.options.Clock()
	r.options.EventHandler(newStartedEvent(startTime))
	for _, cmdController := range cmdControllers {
		cmdController := cmdController
		waitGroup.Add(1)
		go func() {
			semaphore.P(1)
			defer semaphore.V(1)
			defer waitGroup.Done()
			if !cmdController.Run() {
				// best effort to prioritize the interrupt error
				// but this is not deterministic
				err = errCmdFailed
				if r.options.FastFail {
					doneC <- struct{}{}
				}
			}
		}()
	}
	go func() {
		// if everything finishes and there is an interrupt, we could
		// end up not actually returning an error if everything below
		// completes before we context switch to the interrupt goroutine
		waitGroup.Wait()
		doneC <- struct{}{}
	}()
	<-doneC
	for _, cmdController := range cmdControllers {
		cmdController.Kill()
	}
	finishTime := r.options.Clock()
	r.options.EventHandler(newFinishedEvent(finishTime, startTime, err))
	return err
}

type cmdController struct {
	Cmd          *exec.Cmd
	EventHandler EventHandler
	Clock        func() time.Time
	Started      bool
	Finished     bool
	StartTime    time.Time
	Lock         sync.Mutex
}

func newCmdController(cmd *exec.Cmd, eventHandler EventHandler, clock func() time.Time) *cmdController {
	return &cmdController{cmd, eventHandler, clock, false, false, clock(), sync.Mutex{}}
}

// Run returns false on failure that has not been already handled
func (c *cmdController) Run() bool {
	c.Lock.Lock()
	if c.Started || c.Finished {
		c.Lock.Unlock()
		return true
	}
	c.Started = true
	c.StartTime = c.Clock()
	c.EventHandler(newCmdStartedEvent(c.StartTime, c.Cmd))
	if err := c.Cmd.Start(); err != nil {
		finishTime := c.Clock()
		err = fmt.Errorf("command could not start: %s: %v", cmdString(c.Cmd), err)
		c.Finished = true
		c.EventHandler(newCmdFinishedEvent(finishTime, c.Cmd, c.StartTime, err))
		c.Lock.Unlock()
		return false
	}
	c.Lock.Unlock()
	err := c.Cmd.Wait()
	finishTime := c.Clock()
	if err != nil {
		err = fmt.Errorf("command had error: %s: %s", cmdString(c.Cmd), err.Error())
	}
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.Finished {
		return true
	}
	c.Finished = true
	c.EventHandler(newCmdFinishedEvent(finishTime, c.Cmd, c.StartTime, err))
	return err == nil
}

func (c *cmdController) Kill() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if !c.Started {
		c.Started = true
		c.Finished = true
		return
	}
	if c.Finished {
		return
	}
	c.Finished = true
	if c.Cmd.Process != nil {
		err := c.Cmd.Process.Kill()
		finishTime := c.Clock()
		if err != nil {
			err = fmt.Errorf("command had error on kill: %s: %s", cmdString(c.Cmd), err.Error())
		}
		c.EventHandler(newCmdFinishedEvent(finishTime, c.Cmd, c.StartTime, err))
	}
}

type semaphore chan struct{}

func newSemaphore(n int) semaphore {
	if n <= 0 {
		return nil
	}
	s := make(semaphore, n)
	for i := 0; i < n; i++ {
		s <- struct{}{}
	}
	return s
}

func (s semaphore) P(n int) {
	if s == nil {
		return
	}
	for i := 0; i < n; i++ {
		<-s
	}
}

func (s semaphore) V(n int) {
	if s == nil {
		return
	}
	for i := 0; i < n; i++ {
		s <- struct{}{}
	}
}

func cmdString(cmd *exec.Cmd) string {
	return strings.Join(append([]string{cmd.Path}, cmd.Args...), " ")
}

func logEvent(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Print(event.Type())
		return
	}
	log.Print(string(data))
}
