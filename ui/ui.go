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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query"
)

type QueryHandler struct {
	API     api.API
	Backend api.Backend
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Body    interface{} `json:"body,omitempty"`
}

func errorResponse(writer http.ResponseWriter, code int, err error) {
	writer.WriteHeader(code)
	encoded, err := json.MarshalIndent(Response{Success: false, Message: err.Error()}, "", "  ")
	if err != nil {
		writer.Write([]byte("{\"success\":false, \"message\":'failed to encode error message'}"))
		return
	}
	writer.Write(encoded)
}

func bodyResponse(writer http.ResponseWriter, body interface{}) {
	encoded, err := json.MarshalIndent(Response{Success: true, Body: body}, "", "  ")
	if err != nil {
		writer.Write([]byte("{\"success\":false, \"message\":'failed to encode result message'"))
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

	cmd, err := query.Parse(input)
	if err != nil {
		errorResponse(writer, http.StatusBadRequest, err)
		return
	}

	result, err := cmd.Execute(q.Backend, q.API)
	if err != nil {
		errorResponse(writer, http.StatusInternalServerError, err)
		return
	}
	bodyResponse(writer, result)
}

type StaticHandler struct {
	Directory string
}

func (h StaticHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	res := h.Directory + request.URL.Path
	fmt.Printf("res = %s\n", res)
	http.ServeFile(writer, request, res)
}

func Main(apiInstance api.API, backend api.Backend) {
	handler := QueryHandler{
		API:     apiInstance,
		Backend: backend,
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/query", handler)
	here, err := filepath.Abs("")
	if err != nil {
		fmt.Printf("ERROR [%s]\n", err.Error())
		return
	}
	httpMux.Handle("/static/", StaticHandler{here + "/" + filepath.Dir(os.Args[0])})

	server := &http.Server{
		Addr:           ":8080",
		Handler:        httpMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
