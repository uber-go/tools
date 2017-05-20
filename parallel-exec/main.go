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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	"go.uber.org/tools/lib/parallel"

	"github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v2"
)

var (
	flagFastFail          = flag.Bool("fast-fail", false, "Fail on the first command failure")
	flagMaxConcurrentCmds = flag.Int("max-concurrent-cmds", runtime.NumCPU(), "Maximum number of processes to run concurrently, or unlimited if 0")

	errUsage               = fmt.Errorf("Usage: %s configFile", os.Args[0])
	errConfigNil           = errors.New("config is nil")
	errConfigCommandsEmpty = errors.New("config commands is empty")
)

type config struct {
	Commands []string `json:"commands,omitempty" yaml:"commands,omitempty"`
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	flag.Parse()
	if err := do(); err != nil {
		log.Fatal(err)
	}
}

func do() error {
	if len(flag.Args()) != 1 {
		log.Fatal(errUsage.Error())
	}
	config, err := readConfig(flag.Args()[0])
	if err != nil {
		return err
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	log.Println(string(data))
	cmds, err := getCmds(config)
	if err != nil {
		return err
	}
	runnerOptions := []parallel.RunnerOption{parallel.WithMaxConcurrentCmds(*flagMaxConcurrentCmds)}
	if *flagFastFail {
		runnerOptions = append(runnerOptions, parallel.WithFastFail())
	}
	return parallel.NewRunner(runnerOptions...).Run(cmds)
}

func readConfig(configFilePath string) (*config, error) {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	config := &config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func validateConfig(config *config) error {
	if config == nil {
		return errConfigNil
	}
	if len(config.Commands) == 0 {
		return errConfigCommandsEmpty
	}
	return nil
}

func getCmds(config *config) ([]*exec.Cmd, error) {
	var cmds []*exec.Cmd
	for _, line := range config.Commands {
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
		cmds = append(cmds, &exec.Cmd{
			Path:   args[0],
			Args:   args[1:],
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
	}
	return cmds, nil
}
