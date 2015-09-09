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

package mocks

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type FakeHttpClient struct {
	responses map[string]Response
}

type Response struct {
	Body       string
	Delay      time.Duration
	StatusCode int
}

func NewFakeHttpClient() *FakeHttpClient {
	return &FakeHttpClient{
		responses: make(map[string]Response),
	}
}

func (c *FakeHttpClient) SetResponse(url string, r Response) {
	c.responses[url] = r
}

func (c *FakeHttpClient) Get(url string) (*http.Response, error) {
	r, exists := c.responses[url]
	if !exists {
		return nil, fmt.Errorf("Get() received unexpected url %s, mappings: %+v", url, c.responses)
	}

	if r.Delay > 0 {
		time.Sleep(r.Delay)
	}
	resp := http.Response{}
	resp.StatusCode = r.StatusCode
	resp.Body = ioutil.NopCloser(bytes.NewBufferString(r.Body))
	if r.StatusCode/100 == 4 || r.StatusCode/100 == 5 {
		return &resp, errors.New("HTTP Error")
	}
	return &resp, nil
}
