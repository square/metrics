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

package mocks

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type FakeHTTPClient struct {
	responses map[string]Response
}

type Response struct {
	Body       string
	Delay      time.Duration
	StatusCode int
}

func NewFakeHTTPClient() *FakeHTTPClient {
	return &FakeHTTPClient{
		// @@ can inline NewFakeHTTPClient
		responses: make(map[string]Response),
	}
	// @@ &FakeHTTPClient literal escapes to heap
	// @@ make(map[string]Response) escapes to heap
}

func (c *FakeHTTPClient) SetResponse(url string, r Response) {
	// @@ leaking param: url
	// @@ leaking param: r
	c.responses[url] = r
	// @@ can inline (*FakeHTTPClient).SetResponse
}

func (c *FakeHTTPClient) Get(url string) (*http.Response, error) {
	// @@ leaking param: url
	// @@ leaking param content: c
	r, exists := c.responses[url]
	if !exists {
		return nil, fmt.Errorf("Get() received unexpected url %s, mappings: %+v", url, c.responses)
	}
	// @@ url escapes to heap
	// @@ c.responses escapes to heap

	if r.Delay > 0 {
		time.Sleep(r.Delay)
	}
	resp := http.Response{}
	resp.StatusCode = r.StatusCode
	// @@ moved to heap: resp
	resp.Body = ioutil.NopCloser(bytes.NewBufferString(r.Body))
	if r.StatusCode/100 == 4 || r.StatusCode/100 == 5 {
		// @@ inlining call to bytes.NewBufferString
		// @@ inlining call to ioutil.NopCloser
		// @@ composite literal escapes to heap
		// @@ bytes.NewBufferString(r.Body) escapes to heap
		// @@ &bytes.Buffer literal escapes to heap
		// @@ ([]byte)(bytes.s·2) escapes to heap
		// @@ composite literal escapes to heap
		// @@ bytes.NewBufferString(r.Body) escapes to heap
		// @@ &bytes.Buffer literal escapes to heap
		// @@ ([]byte)(bytes.s·2) escapes to heap
		return &resp, errors.New("HTTP Error")
	}
	// @@ inlining call to errors.New
	// @@ &errors.errorString literal escapes to heap
	// @@ &errors.errorString literal escapes to heap
	// @@ &resp escapes to heap
	return &resp, nil
}

// @@ &resp escapes to heap
