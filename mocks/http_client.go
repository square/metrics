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
)

type FakeHttpClient struct {
	responses map[string]string
}

func NewFakeHttpClient() *FakeHttpClient {
	return &FakeHttpClient{
		responses: make(map[string]string),
	}
}

func (c *FakeHttpClient) SetResponse(url, response string) {
	c.responses[url] = response
}

func (c *FakeHttpClient) Get(url string) (*http.Response, error) {
	responseString, exists := c.responses[url]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Get() received unexpected url %s, mappings: %+v", url, c.responses))
	}

	resp := http.Response{}
	resp.Body = ioutil.NopCloser(bytes.NewBufferString(responseString))

	return &resp, nil
}
