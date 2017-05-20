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
	"encoding/json"
	"os/exec"
	"time"
)

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
