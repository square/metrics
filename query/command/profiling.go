// Copyright 2015 - 2016 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"fmt"

	"github.com/square/metrics/inspect"
)

//ProfilingCommand is a Command that also performs profiling actions.
type ProfilingCommand struct {
	Profiler *inspect.Profiler
	Command  Command
}

func NewProfilingCommandWithProfiler(command Command, profiler *inspect.Profiler) Command {
	return ProfilingCommand{
		Profiler: profiler,
		Command:  command,
	}
}

func (cmd ProfilingCommand) Execute(context ExecutionContext) (CommandResult, error) {
	defer cmd.Profiler.Record(fmt.Sprintf("%s.Execute", cmd.Name()))()
	context.Profiler = cmd.Profiler
	result, err := cmd.Command.Execute(context)
	if err != nil {
		return CommandResult{}, err
	}
	profiles := cmd.Profiler.All()
	if len(profiles) != 0 {
		if result.Metadata == nil {
			result.Metadata = map[string]interface{}{}
		}
		result.Metadata["profile"] = profiles
	}
	return result, nil
}

func (cmd ProfilingCommand) Name() string {
	return cmd.Command.Name()
}
