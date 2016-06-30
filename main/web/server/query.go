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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/square/metrics/inspect"
	"github.com/square/metrics/log"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/query/predicate"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	QueryResponse
	Profile []inspect.Profile `json:"profile,omitempty"`
}

type QueryResponse struct {
	Name     string                 `json:"name,omitempty"`
	Body     interface{}            `json:"body,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type queryHandler struct {
	hook    Hook
	context command.ExecutionContext
}

type KeyIs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyIn struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type KeyMatch struct {
	Key   string `json:"key"`
	Regex string `json:"regex"`
}

type Constraint struct {
	Not      *Constraint  `json:"not,omitempty"`
	All      []Constraint `json:"all,omitempty"`
	Any      []Constraint `json:"any,omitempty"`
	KeyIs    *KeyIs       `json:"key_is,omitempty"`
	KeyIn    *KeyIn       `json:"key_in,omitempty"`
	KeyMatch *KeyMatch    `json:"key_match,omitempty"`
}

type singleChecker struct {
	found bool
	name  string
}

func (s *singleChecker) add(success bool, message string) error {
	if !success {
		return nil
	}
	if s.found {
		return fmt.Errorf("already had %q but tried to add %q", s.name, message)
	}
	s.found = true
	s.name = message
	return nil
}

func predicateFromConstraint(c Constraint) (predicate.Predicate, error) {
	only := &singleChecker{}
	if err := only.add(c.Not != nil, "not"); err != nil {
		return nil, err
	}
	if err := only.add(c.All != nil, "all"); err != nil {
		return nil, err
	}
	if err := only.add(c.Any != nil, "any"); err != nil {
		return nil, err
	}
	if err := only.add(c.KeyIs != nil, "key_is"); err != nil {
		return nil, err
	}
	if err := only.add(c.KeyIn != nil, "key_in"); err != nil {
		return nil, err
	}
	if err := only.add(c.KeyMatch != nil, "key_match"); err != nil {
		return nil, err
	}
	if !only.found {
		return nil, fmt.Errorf("constraint has no contents")
	}
	switch only.name {
	case "not":
		child, err := predicateFromConstraint(*c.Not)
		if err != nil {
			return nil, err
		}
		return predicate.NotPredicate{child}, nil
	case "all":
		children := make([]predicate.Predicate, len(c.All))
		for i, arg := range c.All {
			child, err := predicateFromConstraint(arg)
			if err != nil {
				return nil, err
			}
			children[i] = child
		}
		return predicate.AndPredicate{children}, nil
	case "any":
		children := make([]predicate.Predicate, len(c.Any))
		for i, arg := range c.Any {
			child, err := predicateFromConstraint(arg)
			if err != nil {
				return nil, err
			}
			children[i] = child
		}
		return predicate.OrPredicate{children}, nil
	case "key_is":
		if c.KeyIs.Key == "" {
			return nil, fmt.Errorf(`key is given no value in "key_is" constraint`)
		}
		return predicate.ListMatcher{
			Tag:    c.KeyIs.Key,
			Values: []string{c.KeyIs.Value},
		}, nil
	case "key_in":
		if c.KeyIn.Key == "" {
			return nil, fmt.Errorf(`key is given no value in "key_in" constraint`)
		}
		return predicate.ListMatcher{
			Tag:    c.KeyIn.Key,
			Values: c.KeyIn.Values,
		}, nil
	case "key_match":
		if c.KeyMatch.Key == "" {
			return nil, fmt.Errorf(`key is given no value in "key_match" constraint`)
		}
		regex, err := regexp.Compile(c.KeyMatch.Regex)
		if err != nil {
			return nil, err
		}
		return predicate.RegexMatcher{
			Tag:   c.KeyMatch.Key,
			Regex: regex,
		}, nil
	default:
		panic(fmt.Sprintf("internal error: unknown constraint name: %q", only.name))
	}
}

type QueryForm struct {
	Input       string      `query:"query" json:"query"`     // query to execute.
	Profile     bool        `query:"profile" json:"profile"` // if true, then profile information will be exposed to the user.
	Constraints *Constraint `query:"-" json:"where"`
}

func (q queryHandler) process(profiler *inspect.Profiler, parsedForm QueryForm) (QueryResponse, error) {
	log.Infof("INPUT: %+v\n", parsedForm)
	var rawCommand command.Command
	var err error
	profiler.Do("Parsing Query", func() {
		rawCommand, err = parser.Parse(parsedForm.Input)
	})
	if err != nil {
		return QueryResponse{}, err
	}

	context := q.context

	if parsedForm.Constraints != nil {
		predicate, err := predicateFromConstraint(*parsedForm.Constraints)
		if err != nil {
			return QueryResponse{}, err
		}
		context.AdditionalConstraints = predicate // Attach the predicate to the context.
	}

	profiledCommand := command.NewProfilingCommandWithProfiler(rawCommand, profiler)

	result := command.CommandResult{}
	profiler.Do("Total Execution", func() {
		result, err = profiledCommand.Execute(context)
	})
	if err != nil {
		return QueryResponse{}, err
	}

	return QueryResponse{
		Body:     result.Body,
		Metadata: result.Metadata,
		Name:     profiledCommand.Name(),
	}, nil
}

// ErrorHTTP indicates that an error should override the return code.
type HTTPError interface {
	error
	ErrorCode() int
}

func (q queryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	profiler := inspect.New()

	queryForm := QueryForm{}

	switch request.Header.Get("Content-Type") {
	case "application/json": // assume the body is a JSON request
		if err := json.NewDecoder(request.Body).Decode(&queryForm); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(encodeError(err))
		}
	default: // use the form parameters
		if err := request.ParseForm(); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(encodeError(err))
			return
		}
		parseStruct(request.Form, &queryForm)
	}

	// "process" does the hard work for the handler, but doesn't touch the HTTP details.
	responseMessage, err := q.process(profiler, queryForm)
	if err != nil {
		if errHTTP, ok := err.(HTTPError); ok {
			// If an HTTPError is returned, then we use its reported code instead of
			// StatusBadRequest. This can be used to identify errors as 500s instead
			// of always blaming the client.
			writer.WriteHeader(errHTTP.ErrorCode())
		} else {
			writer.WriteHeader(http.StatusBadRequest)
		}
		writer.Write(encodeError(err))
		return
	}

	responseJSON := Response{
		Success:       true,
		QueryResponse: responseMessage,
	}

	if showProfile, _ := strconv.ParseBool(request.Form.Get("profile")); showProfile {
		responseJSON.Profile = profiler.All()
	}

	if q.hook.OnQuery != nil {
		go func() {
			// Send the profiler along this way.
			q.hook.OnQuery <- profiler
		}()
	}

	pretty, _ := strconv.ParseBool(request.Form.Get("pretty")) // If it's absent, default to false.

	var encoded []byte
	if pretty {
		encoded, err = json.MarshalIndent(responseJSON, "", "  ")
	} else {
		encoded, err = json.Marshal(responseJSON)
	}
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(encodeError(err))
		return
	}

	writer.Write(encoded)
}
