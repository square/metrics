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

package query

import (
	"fmt"

	"github.com/square/metrics/query/aggregate"
)

type MetricFunction struct {
	Name        string
	MinArgument int
	MaxArgument int
	Groups      bool // Whether the function allows a 'group by' clause.
	Compute     func(EvaluationContext, []Expression, []string) (value, error)
}

func (f MetricFunction) Evaluate(
	context EvaluationContext,
	arguments []Expression,
	groupBy []string,
) (value, error) {
	// preprocessing
	length := len(arguments)
	if length < f.MinArgument || (f.MaxArgument != -1 && f.MaxArgument < length) {
		return nil, ArgumentLengthError{f.Name, f.MinArgument, f.MaxArgument, length}
	}
	if len(groupBy) > 0 && !f.Groups {
		return nil, fmt.Errorf("function %s doesn't allow a group-by clause", f.Name)
	}
	return f.Compute(context, arguments, groupBy)
}

// MakeAggregateMetricFunction takes a named aggregating function `[float64] => float64` and makes it into a MetricFunction.
func MakeAggregateMetricFunction(name string, aggregator func([]float64) float64) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: 1,
		MaxArgument: 1,
		Compute: func(context EvaluationContext, args []Expression, groups []string) (value, error) {
			argument := args[0]
			value, err := argument.Evaluate(context)
			if err != nil {
				return nil, err
			}
			seriesList, err := value.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			result := aggregate.AggregateBy(seriesList, aggregator, groups)
			return seriesListValue(result), nil
		},
	}
}

// MakeTransformMetircFunction takes a named transforming function `[float64], [value] => [float64]` and makes it into a MetricFunction.
func MakeTransformMetricFunction(name string, parameterCount int, transformer func([]float64, []value, float64) ([]float64, error)) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: parameterCount + 1,
		MaxArgument: parameterCount + 1,
		Groups:      true,
		Compute: func(context EvaluationContext, args []Expression, groups []string) (value, error) {
			// ApplyTransform(list api.SeriesList, transform transform, parameters []value) (api.SeriesList, error)
			listValue, err := args[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			list, err := listValue.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			parameters := make([]value, parameterCount)
			for i := range parameters {
				parameters[i], err = args[i+1].Evaluate(context)
				if err != nil {
					return nil, err
				}
			}
			result, err := ApplyTransform(list, transformer, parameters)
			if err != nil {
				return nil, err
			}
			return seriesListValue(result), nil
		},
	}
}

var TimeshiftFunction = MetricFunction{
	Name:        "timeshift",
	MinArgument: 2,
	MaxArgument: 2,
	Compute: func(context EvaluationContext, arguments []Expression, groups []string) (value, error) {
		value, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		millis, err := toDuration(value)
		if err != nil {
			return nil, err
		}
		newContext := context
		newContext.Timerange = newContext.Timerange.Shift(millis)

		result, err := arguments[0].Evaluate(newContext)
		if err != nil {
			return nil, err
		}

		if seriesValue, ok := result.(seriesListValue); ok {
			seriesValue.Timerange = context.Timerange
			return seriesValue, nil
		}
		return result, nil

	},
}
