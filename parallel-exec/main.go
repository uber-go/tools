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

package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/mattn/go-shellwords"

	"go.uber.org/tools/parallel-exec/lib"
)

var (
	flagFastFail          = flag.Bool("fast-fail", false, "Fail on the first command failure")
	flagMaxConcurrentCmds = flag.Int("max-concurrent-cmds", runtime.NumCPU(), "Maximum number of processes to run concurrently, or unlimited if 0")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	flag.Parse()
	if err := do(); err != nil {
		log.Fatal(err)
	}
}

func do() error {
	cmds, err := getCmds()
	if err != nil {
		return err
	}
	runnerOptions := []lib.RunnerOption{lib.WithMaxConcurrentCmds(*flagMaxConcurrentCmds)}
	if *flagFastFail {
		runnerOptions = append(runnerOptions, lib.WithFastFail())
	}
	return lib.NewRunner(runnerOptions...).Run(cmds)
}

func getCmds() ([]*exec.Cmd, error) {
	var cmds []*exec.Cmd
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		args, err := shellwords.Parse(line)
		if err != nil {
			return nil, err
		}
		// could happen if args = "$FOO" and FOO is not set
		if len(args) == 0 {
			continue
		}
		cmds = append(cmds, exec.Command(args[0], args[1:]...))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cmds, nil
}
