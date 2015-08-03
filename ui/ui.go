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
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/square/metrics/inspect"
	"github.com/square/metrics/log"
	_ "github.com/square/metrics/main/static/common"  // ensure that the static files are included.
	_ "github.com/square/metrics/main/static/example" // ensure that the static files are included.
	"github.com/square/metrics/query"
)

var failedMessage []byte

func init() {
	var err error
	failedMessage, err = json.MarshalIndent(response{Success: false, Message: "Failed to encode the result message."}, "", "  ")
	if err != nil {
		panic(err.Error())
	}
}

// tokenHandler exposes all the tokens available in the system for the autocomplete.
type tokenHandler struct {
	hook    Hook
	context query.ExecutionContext
}

type queryHandler struct {
	hook    Hook
	context query.ExecutionContext
}

// generic response functions
// --------------------------
func commonResponse(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
}
func errorResponse(writer http.ResponseWriter, code int, err error) {
	commonResponse(writer)
	writer.WriteHeader(code)
	encoded, err := json.MarshalIndent(response{Success: false, Message: err.Error()}, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(failedMessage)
		return
	}
	writer.Write(encoded)
}

func bodyResponse(writer http.ResponseWriter, response response) {
	commonResponse(writer)
	response.Success = true
	encoded, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(failedMessage)
		return
	}
	writer.Write(encoded)
}

// parsing functions
// -----------------

type queryForm struct {
	input   string // query to execute.
	profile bool   // if true, then profile information will be exposed to the user.
}

func parseBool(input string, defaultValue bool) bool {
	value, err := strconv.ParseBool(input)
	if err != nil {
		return defaultValue
	}
	return value
}

func parseQueryForm(request *http.Request) (form queryForm) {
	form.input = request.Form.Get("query")
	form.profile = parseBool(request.Form.Get("profile"), false)
	return
}

func convertProfile(profiler *inspect.Profiler) []profileJSON {
	profiles := profiler.All()
	result := make([]profileJSON, len(profiles))
	for i, p := range profiles {
		result[i] = profileJSON{
			Name:   p.Name(),
			Start:  p.Start().UnixNano() / int64(time.Millisecond),
			Finish: p.Finish().UnixNano() / int64(time.Millisecond),
		}
	}
	return result
}

func (h tokenHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := make(map[string]interface{}) // map to array-like types.
	// extract out all the possible tokens
	// 1. keywords
	// 2. functions
	// 3. identifiers
	body["functions"] = h.context.Registry.All()
	metrics, err := h.context.API.GetAllMetrics()
	if err != nil {
		errorResponse(writer, http.StatusInternalServerError, err)
		return
	} else {
		body["metrics"] = metrics
	}
	response := response{
		Body: body,
	}
	bodyResponse(writer, response)
}

func (q queryHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
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
	response := response{
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

type staticHandler struct {
	Web       string
	Directory string
}

func (h staticHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	web := h.Web
	reqPath := request.URL.Path
	for len(web) > 0 && len(reqPath) > 0 && web[0] == reqPath[0] {
		web = web[1:]
		reqPath = reqPath[1:]
	}
	res := path.Join(h.Directory, reqPath)
	log.Infof("url.path=%s, resource=%s\n", reqPath, res)
	http.ServeFile(writer, request, res)
}

type singleStaticHandler struct {
	Path string
}

func (h singleStaticHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, h.Path)
}

func NewMux(config Config, context query.ExecutionContext, hook Hook) *http.ServeMux {
	// Wrap the given API and Backend in their Profiling counterparts.
	httpMux := http.NewServeMux()
	for web, file := range config.Routes {
		// Attach each web:file pair to an appropriate handle
		httpMux.Handle(web, staticHandler{web, file})
	}
	/*httpMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/ui", http.StatusTemporaryRedirect)
	})*/
	httpMux.Handle("/query", queryHandler{
		context: context,
		hook:    hook,
	})
	httpMux.Handle("/token", tokenHandler{
		context: context,
		hook:    hook,
	})
	return httpMux
}
