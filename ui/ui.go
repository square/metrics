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
	"log"
	"net/http"
	"time"

	"github.com/square/"
)

func queryHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello, world!"))
}

func Main(apiInstance api.API, backend api.Backend) {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/query", queryHandler)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        httpMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
