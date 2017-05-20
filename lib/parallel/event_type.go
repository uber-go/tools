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
	"fmt"
	"strconv"
	"strings"
)

const (
	// EventTypeStarted says that the runner started.
	EventTypeStarted EventType = iota + 1
	// EventTypeCmdStarted says that a command started.
	EventTypeCmdStarted
	// EventTypeCmdFinished says that a command finished.
	EventTypeCmdFinished
	// EventTypeFinished says that the runner finished.
	EventTypeFinished
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

// MarshalJSON marshals the EventType to JSON.
func (e EventType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + e.String() + `"`), nil
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
