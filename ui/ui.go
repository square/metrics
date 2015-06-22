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
	"time"

	"github.com/square/metrics/log"
	_ "github.com/square/metrics/main/static" // ensure that the static files are included.
	"github.com/square/metrics/query"
)

type Config struct {
	Port      int    `yaml:"port"`
	Timeout   int    `yaml:"timeout"`
	StaticDir string `yaml:"static_dir"`
}

type QueryHandler struct {
	context query.ExecutionContext
}

type Response struct {
	Success bool        `json:"success"`
	Name    string      `json:"name,omitempty"`
	Message string      `json:"message,omitempty"`
	Body    interface{} `json:"body,omitempty"`
}

func errorResponse(writer http.ResponseWriter, code int, err error) {
	writer.WriteHeader(code)
	encoded, err := json.MarshalIndent(Response{Success: false, Message: err.Error()}, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"success\":false, \"message\":\"failed to encode error message\"}"))
		return
	}
	writer.Write(encoded)
}

func bodyResponse(writer http.ResponseWriter, body interface{}, name string) {
	encoded, err := json.MarshalIndent(Response{Success: true, Name: name, Body: body}, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"success\":false, \"message\":\"failed to encode result message\"}"))
		return
	}
	writer.Write(encoded)
}

func (q QueryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		errorResponse(writer, http.StatusBadRequest, err)
		return
	}
	input := request.Form.Get("query")
	log.Infof("INPUT: %+v\n", input)

	cmd, err := query.Parse(input)
	if err != nil {
		errorResponse(writer, http.StatusBadRequest, err)
		return
	}

	result, err := cmd.Execute(q.context)
	if err != nil {
		errorResponse(writer, http.StatusInternalServerError, err)
		return
	}
	bodyResponse(writer, result, cmd.Name())
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

func NewMux(config Config, context query.ExecutionContext) *http.ServeMux {
	httpMux := http.NewServeMux()
	httpMux.Handle("/query", QueryHandler{
		context: context,
	})
	staticPath := "/static/"
	httpMux.Handle(staticPath, StaticHandler{StaticPath: staticPath, Directory: config.StaticDir})
	return httpMux
}

func Main(config Config, context query.ExecutionContext) {
	httpMux := NewMux(config, context)

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
