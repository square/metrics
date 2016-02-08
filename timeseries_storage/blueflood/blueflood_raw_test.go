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

package blueflood

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/testing_support/mocks"
)

func TestIncludeRawPayload(t *testing.T) {
	now := time.Unix(1438734300000, 0)

	baseTime := now.Unix() * 1000
	timeSource := func() time.Time { return now }

	queryTimerange, err := api.NewSnappedTimerange(
		int64(baseTime)-300*1000*10, // 50 minutes ago
		int64(baseTime)-300*1000*4,  // 20 minutes ago
		300*1000,                    // 5 minute resolution
	)

	// The queries have to be relative to "now"
	defaultClientConfig := Config{
		BaseUrl:               "https://blueflood.url",
		TenantId:              "square",
		Ttls:                  make(map[string]int64),
		Timeout:               time.Millisecond,
		FullResolutionOverlap: 14400,
		TimeSource:            timeSource,
	}

	regularQueryURL := fmt.Sprintf(
		"https://blueflood.url/v2.0/square/views/some.key.value?from=%d&resolution=MIN5&select=numPoints%%2Caverage&to=%d",
		queryTimerange.Start(),
		queryTimerange.End()+queryTimerange.ResolutionMillis(),
	)

	fullResolutionQueryURL := fmt.Sprintf(
		"https://blueflood.url/v2.0/square/views/some.key.value?from=%d&resolution=FULL&select=numPoints%%2Caverage&to=%d",
		queryTimerange.Start(),
		queryTimerange.End()+queryTimerange.ResolutionMillis(),
	)

	regularResponse := fmt.Sprintf(`{
	  "unit": "unknown",
	  "values": [
	    {
	      "numPoints": 28,
	      "timestamp": %d,
	      "average": 100
	    },
	    {
	      "numPoints": 29,
	      "timestamp": %d,
	      "average": 142
	    },
	    {
	      "numPoints": 27,
	      "timestamp": %d,
	      "average": 138
	    },
	    {
	      "numPoints": 28,
	      "timestamp": %d,
	      "average": 182
	    }
	  ],
	  "metadata": {
	    "limit": null,
	    "next_href": null,
	    "count": 4,
	    "marker": null
	  }
	}`,
		baseTime-300*1000*10, // 50 minutes ago
		baseTime-300*1000*9,  // 45 minutes ago
		baseTime-300*1000*8,  // 40 minutes ago
		baseTime-300*1000*7,  // 35 minutes ago
	)

	fakeHttpClient := mocks.NewFakeHttpClient()
	fakeHttpClient.SetResponse(regularQueryURL, mocks.Response{regularResponse, 0, http.StatusOK})
	// Note that the following is cheating, since it's not a reasonable response for that query.
	// It still tests this aspect of the raw_test, however, so it's kept.
	fakeHttpClient.SetResponse(fullResolutionQueryURL, mocks.Response{regularResponse, 0, http.StatusOK})
	defaultClientConfig.HttpClient = fakeHttpClient
	defaultClientConfig.TimeSource = timeSource

	b := NewBlueflood(defaultClientConfig)
	if err != nil {
		t.Fatalf("timerange error: %s", err.Error())
	}

	userConfig := api.UserSpecifiableConfig{
		IncludeRawData: true,
	}

	evaluationNotes := &inspect.EvaluationNotes{}

	if _, err := b.FetchSingleTimeseries(api.FetchTimeseriesRequest{
		Metric:                "some.key.value",
		SampleMethod:          api.SampleMean,
		Timerange:             queryTimerange,
		Cancellable:           api.NewCancellable(),
		EvaluationNotes:       evaluationNotes,
		UserSpecifiableConfig: userConfig,
	}); err != nil {
		t.Fatalf("Expected success, but got error: %s", err.Error())
	}

	notes := evaluationNotes.Notes()

	if len(notes) != 2 {
		t.Fatalf("Expected 2 evaluation notes: for this series, one for full resolution and one for query resolution, but got %d: %+v", len(notes), notes)
	}
	if expected := "Blueflood (query resolution) some.key.value: " + regularResponse; notes[0] != expected {
		t.Errorf("Didn't fill in notes[0] correctly, got\n%s\n but expected \n%s\n", notes[0], expected)
	}
	if expected := "Blueflood (full resolution) some.key.value: " + regularResponse; notes[1] != expected {
		t.Errorf("Didn't fill in notes[1] correctly, got\n%s\n but expected \n%s\n", notes[1], expected)
	}

}
