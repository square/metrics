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

package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"

	"github.com/square/metrics/log"
)

func encodeError(err error) []byte {
	encoded, err2 := json.MarshalIndent(Response{
		Success: false,
		Message: err.Error(),
	}, "", "  ")
	if err2 == nil {
		return encoded
	}
	log.Errorf("In query handler: json.Marshal(%+v) returned %+v", err, err2)
	return []byte(`{"success":false, "error": "internal server error while marshalling error message"}`)
}

// parsing functions
// -----------------

type singleStaticHandler struct {
	Path string
	File string
}

func (h singleStaticHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, path.Join(h.Path, h.File))
}

func parseStruct(form url.Values, target interface{}) {
	targetPointer := reflect.ValueOf(target)
	if targetPointer.Type().Kind() != reflect.Ptr {
		panic("Cannot parseStruct into non-pointer")
	}
	targetValue := targetPointer.Elem()
	targetType := targetValue.Type()
	if targetType.Kind() != reflect.Struct {
		panic("Cannot parseStruct into pointer to non-struct")
	}
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		name := field.Name
		if alternate := field.Tag.Get("query"); alternate != "" {
			name = alternate
		}
		if field.Tag.Get("query") == "-" {
			continue // Skip the query
		}
		keyValue := form.Get(name)
		if field.Type.Kind() == reflect.String {
			targetValue.Field(i).Set(reflect.ValueOf(keyValue))
			continue
		}
		if field.Type.Kind() == reflect.Bool {
			store, err := strconv.ParseBool(form.Get(name))
			if err != nil {
				continue // Do nothing
			}
			targetValue.Field(i).Set(reflect.ValueOf(store))
			continue
		}
		if field.Tag.Get("query_kind") == "json" {
			json.Unmarshal([]byte(form.Get(name)), targetValue.Field(i).Addr().Interface())
			continue
		}
		panic(fmt.Sprintf("parseStruct cannot handle field %+v (of an unimplemented type). Consider adding the tag `query_kind:\"json\"`", field))
	}
}
