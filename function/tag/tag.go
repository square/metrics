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

package tag

import (
	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// dropTagSeries returns a copy of the timeseries where the given `dropTag` has been removed from its TagSet.
func dropTagSeries(series api.Timeseries, dropTag string) api.Timeseries {
	tagSet := api.NewTagSet()
	for tag, val := range series.TagSet {
		if tag != dropTag {
			tagSet[tag] = val
		}
	}
	series.TagSet = tagSet
	return series
}

// DropTag returns a copy of the series list where the given `tag` has been removed from all timeseries.
func DropTag(list api.SeriesList, tag string) api.SeriesList {
	series := make([]api.Timeseries, len(list.Series))
	for i := range series {
		series[i] = dropTagSeries(list.Series[i], tag)
	}
	return api.SeriesList{
		series,
	}
}

// setTagSeries returns a copy of the timeseries where the given `newTag` has been set to `newValue`, or added if it wasn't present.
func setTagSeries(series api.Timeseries, newTag string, newValue string) api.Timeseries {
	tagSet := api.NewTagSet()
	for tag, val := range series.TagSet {
		tagSet[tag] = val
	}
	tagSet[newTag] = newValue
	series.TagSet = tagSet
	return series
}

// SetTag returns a copy of the series list where `tag` has been assigned to `value` for every timeseries in the list.
func SetTag(list api.SeriesList, tag string, value string) api.SeriesList {
	series := make([]api.Timeseries, len(list.Series))
	for i := range series {
		series[i] = setTagSeries(list.Series[i], tag, value)
	}
	return api.SeriesList{
		series,
	}
}

// DropFunction wraps up DropTag into a MetricFunction called "tag.drop"
var DropFunction = function.MetricFunction{
	Name:         "tag.drop",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		list, err := function.EvaluateToSeriesList(arguments[0], context)
		if err != nil {
			return nil, err
		}
		dropTag, err := function.EvaluateToString(arguments[1], context)
		if err != nil {
			return nil, err
		}
		// Drop the tag from the list.
		return DropTag(list, dropTag), nil
	},
}

// SetFunction wraps up SetTag into a MetricFunction called "tag.set"
var SetFunction = function.MetricFunction{
	Name:         "tag.set",
	MinArguments: 3,
	MaxArguments: 3,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		list, err := function.EvaluateToSeriesList(arguments[0], context)
		if err != nil {
			return nil, err
		}
		tag, err := function.EvaluateToString(arguments[1], context)
		if err != nil {
			return nil, err
		}
		set, err := function.EvaluateToString(arguments[2], context)
		if err != nil {
			return nil, err
		}
		// Set the tag for the list:
		return SetTag(list, tag, set), nil
	},
}
