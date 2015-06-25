// Copyright 2015 Square Inc.
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

package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/square/metrics/inspect"
	"github.com/square/metrics/log"
	_ "github.com/square/metrics/main/static" // ensure that the static files are included.
	"github.com/square/metrics/query"
)

var failedMessage []byte

func init() {
	var err error
	failedMessage, err = json.MarshalIndent(Response{Success: false, Message: "Failed to encode the result message."}, "", "  ")
	if err != nil {
		panic(err.Error())
	}
}

type Config struct {
	Port      int    `yaml:"port"`
	Timeout   int    `yaml:"timeout"`
	StaticDir string `yaml:"static_dir"`
}

type Hook struct {
	OnQuery chan<- *inspect.Profiler
}

type QueryHandler struct {
	hook    Hook
	context query.ExecutionContext
}

type Response struct {
	Success bool          `json:"success"`
	Name    string        `json:"name,omitempty"`
	Message string        `json:"message,omitempty"`
	Body    interface{}   `json:"body,omitempty"`
	Profile []ProfileJSON `json:"body,omitempty"`
}

type ProfileJSON struct {
	Name   string `json:"name"`
	Start  int64  `json:"start"`  // ms since Unix epoch
	Finish int64  `json:"finish"` // ms since Unix epoch
}

func errorResponse(writer http.ResponseWriter, code int, err error) {
	writer.WriteHeader(code)
	encoded, err := json.MarshalIndent(Response{Success: false, Message: err.Error()}, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(failedMessage)
		return
	}
	writer.Write(encoded)
}

func bodyResponse(writer http.ResponseWriter, response Response) {
	response.Success = true
	encoded, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(failedMessage)
		return
	}
	writer.Write(encoded)
}

type QueryForm struct {
	input   string
	profile bool
}

func parseBool(input string, defaultValue bool) bool {
	value, err := strconv.ParseBool(input)
	if err != nil {
		return defaultValue
	}
	return value
}

func parseQueryForm(request *http.Request) (form QueryForm) {
	form.input = request.Form.Get("query")
	form.profile = parseBool(request.Form.Get("profile"), false)
	return
}

func convertProfile(profiler *inspect.Profiler) []ProfileJSON {
	profiles := profiler.All()
	result := make([]ProfileJSON, len(profiles))
	for i, p := range profiles {
		result[i] = ProfileJSON{
			Name:   p.Name(),
			Start:  p.Start().UnixNano() / int64(time.Millisecond),
			Finish: p.Finish().UnixNano() / int64(time.Millisecond),
		}
	}
	return result
}

func (q QueryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		errorResponse(writer, http.StatusBadRequest, err)
		return
	}
	parsedForm := parseQueryForm(request)
	log.Infof("INPUT: %+v\n", parsedForm)

	cmd, err := query.Parse(parsedForm.input)
	if err != nil {
		errorResponse(writer, http.StatusBadRequest, err)
		return
	}

	cmd, profiler := query.NewProfilingCommand(cmd)
	result, err := cmd.Execute(q.context)
	if err != nil {
		errorResponse(writer, http.StatusInternalServerError, err)
		return
	}
	log.Infof("Profiler results: %+v", profiler.All())
	response := Response{
		Body: result,
		Name: cmd.Name(),
	}
	if parsedForm.profile {
		response.Profile = convertProfile(profiler)
	}
	bodyResponse(writer, response)
	if q.hook.OnQuery != nil {
		q.hook.OnQuery <- profiler
	}
}

type StaticHandler struct {
	Directory  string
	StaticPath string
}

func (h StaticHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	res := h.Directory + request.URL.Path[len(h.StaticPath):]
	log.Infof("url.path = %s\n", request.URL.Path)
	log.Infof("res = %s\n", res)
	http.ServeFile(writer, request, res)
}

func NewMux(config Config, context query.ExecutionContext, hook Hook) *http.ServeMux {
	// Wrap the given API and Backend in their Profiling counterparts.

	httpMux := http.NewServeMux()
	httpMux.Handle("/query", QueryHandler{
		context: context,
		hook:    hook,
	})
	staticPath := "/static/"
	httpMux.Handle(staticPath, StaticHandler{StaticPath: staticPath, Directory: config.StaticDir})
	return httpMux
}

func Main(config Config, context query.ExecutionContext) {
	httpMux := NewMux(config, context, Hook{})

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		Handler:        httpMux,
		ReadTimeout:    time.Duration(config.Timeout) * time.Second,
		WriteTimeout:   time.Duration(config.Timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Infof(err.Error())
	}
}
