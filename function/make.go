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

package function

import (
	"fmt"
	"reflect"
	"time"

	"github.com/square/metrics/api"
)

var stringType = reflect.TypeOf("")
var scalarType = reflect.TypeOf(float64(0.0))
var durationType = reflect.TypeOf(time.Duration(0))
var timeseriesType = reflect.TypeOf(api.SeriesList{})
var valueType = reflect.TypeOf((*Value)(nil)).Elem()
var expressionType = reflect.TypeOf((*Expression)(nil)).Elem()
var groupsType = reflect.TypeOf(Groups{})
var contextType = reflect.TypeOf(EvaluationContext{})
var timerangeType = reflect.TypeOf(api.Timerange{})

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// MakeFunction is a convenient way to use type-safe functions to
// construct MetricFunctions without manually checking parameters.
func MakeFunction(name string, function interface{}) MetricFunction {
	funcValue := reflect.ValueOf(function)
	if funcValue.Kind() != reflect.Func {
		panic("MetricFunction expects a function as input.")
	}
	funcType := funcValue.Type()
	if funcType.NumOut() == 0 {
		panic("MetricFunction's argument function must return a value.")
	}
	if funcType.NumOut() > 2 {
		panic("MetricFunction's argument function must return at most two values.")
	}
	// TODO: allow subtypes
	if !funcType.Out(0).ConvertibleTo(valueType) {
		panic("MetricFunction's argument function's first return type must be assignable to `function.Value`")
	}
	if funcType.NumOut() == 2 && !funcType.Out(1).ConvertibleTo(errorType) {
		panic("MetricFunction's argument function's second return type must be `error`.")
	}

	formalArgumentCount := 0
	allowsGroupBy := false
	for i := 0; i < funcType.NumIn(); i++ {
		argType := funcType.In(i)
		if argType == contextType || argType == timerangeType {
			// It asks for context (or part of it).
			continue
		}
		if argType == groupsType {
			// It asks for groups.
			allowsGroupBy = true
			continue
		}
		formalArgumentCount++ // The next thing is an actual argument.
		if argType == stringType || argType == scalarType || argType == durationType || argType == timeseriesType {
			// Everything is okay
			continue
		}
		if argType == valueType {
			// It's untyped.
			continue
		}
		if argType == expressionType {
			// It's lazy.
			continue
		}
		// TODO: handle optional arguments
		panic(fmt.Sprintf("MetricFunction function argument asks for unsupported type: cannot supply argument %d of type %+v", i, argType))
	}
	// We've checked that everything is sound.
	return MetricFunction{
		Name:          name,
		MinArguments:  formalArgumentCount,
		MaxArguments:  formalArgumentCount,
		AllowsGroupBy: allowsGroupBy,
		Compute: func(context EvaluationContext, arguments []Expression, groups Groups) (Value, error) {
			argValues := make([]reflect.Value, funcType.NumIn())
			// TODO: evaluate in parallel where possible.
			formalArgument := 0
			for i := 0; i < funcType.NumIn(); i++ {
				argType := funcType.In(i)
				if argType == contextType {
					argValues[i] = reflect.ValueOf(context)
					continue
				}
				if argType == timerangeType {
					argValues[i] = reflect.ValueOf(context.Timerange)
					continue
				}
				if argType == groupsType {
					argValues[i] = reflect.ValueOf(groups)
					continue
				}
				if argType == stringType {
					stringArg, err := EvaluateToString(arguments[formalArgument], context)
					if err != nil {
						return nil, err
					}
					argValues[i] = reflect.ValueOf(stringArg)
					formalArgument++
					continue
				}
				if argType == scalarType {
					scalarArg, err := EvaluateToScalar(arguments[formalArgument], context)
					if err != nil {
						return nil, err
					}
					argValues[i] = reflect.ValueOf(scalarArg)
					formalArgument++
					continue
				}
				if argType == durationType {
					durationArg, err := EvaluateToDuration(arguments[formalArgument], context)
					if err != nil {
						return nil, err
					}
					argValues[i] = reflect.ValueOf(durationArg)
					formalArgument++
					continue
				}
				if argType == timeseriesType {
					timeseriesArg, err := EvaluateToSeriesList(arguments[formalArgument], context)
					if err != nil {
						return nil, err
					}
					argValues[i] = reflect.ValueOf(timeseriesArg)
					formalArgument++
					continue
				}
				if argType == valueType {
					valueArg, err := arguments[formalArgument].Evaluate(context)
					if err != nil {
						return nil, err
					}
					argValues[i] = reflect.ValueOf(valueArg)
					formalArgument++
					continue
				}
				if argType == expressionType {
					argValues[i] = reflect.ValueOf(arguments[formalArgument])
					formalArgument++
					continue
				}
				panic(fmt.Sprintf("Argument to MakeFunction requests invalid type %+v", argType))
			}
			output := funcValue.Call(argValues)
			if len(output) == 1 {
				return output[0].Interface().(Value), nil
			}
			if len(output) == 2 {
				valueI, errI := output[0].Interface(), output[1].Interface()
				if errI != nil {
					return nil, errI.(error)
				}
				return valueI.(Value), nil
			}
			panic("MakeFunction built with function that doesn't return 2 things.")
		},
	}

}
