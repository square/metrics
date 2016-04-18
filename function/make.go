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
	if !funcType.Out(0).ConvertibleTo(valueType) {
		panic("MetricFunction's argument function's first return type must be convertible to `function.Value`.")
	}
	if funcType.NumOut() == 2 && !funcType.Out(1).ConvertibleTo(errorType) {
		panic("MakeFunction's argument function's second return type must convertible be `error`.")
	}

	formalArgumentCount := 0
	optionalArgumentCount := 0
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
		if argType.Kind() == reflect.Ptr {
			// Then this argument is optional (but everything else is the same).
			argType = argType.Elem()
			optionalArgumentCount++
		} else {
			// If an optional argument exists, this one must be optional too:
			if optionalArgumentCount > 0 {
				panic("MakeFunction's argument function has non-optional formal parameter after optional formal parameter.")
			}
		}
		formalArgumentCount++ // The next thing is an actual argument.
		switch argType {
		case stringType, scalarType, durationType, timeseriesType, valueType, expressionType:
			// Do nothing: these are all okay.
		default:
			panic(fmt.Sprintf("MetricFunction function argument asks for unsupported type: cannot supply argument %d of type %+v.", i, argType))
		}
	}
	// We've checked that everything is sound.
	return MetricFunction{
		Name:          name,
		MinArguments:  formalArgumentCount - optionalArgumentCount,
		MaxArguments:  formalArgumentCount,
		AllowsGroupBy: allowsGroupBy,
		Compute: func(context EvaluationContext, arguments []Expression, groups Groups) (Value, error) {
			argValues := make([]reflect.Value, funcType.NumIn())
			// TODO: evaluate in parallel where possible.

			formalArgument := 0
			nextArgument := func() Expression {
				if formalArgument >= len(arguments) {
					return nil
				}
				arg := arguments[formalArgument]
				formalArgument++
				return arg
			}

			evalTo := func(expression Expression, result reflect.Type) (interface{}, error) {
				value, err := expression.Evaluate(context)
				if err != nil {
					return reflect.Value{}, err
				}
				switch result {
				case stringType:
					return value.ToString("TODO: what goes here?")
				case scalarType:
					return value.ToScalar("TODO: what goes here?")
				case durationType:
					return value.ToDuration("TODO: what goes here?")
				case timeseriesType:
					return value.ToSeriesList(context.Timerange)
				case valueType:
					return value, nil
				}
				panic("Unknown type!!!")
			}

			// TODO: get pointers working for optional arguments
			for i := 0; i < funcType.NumIn(); i++ {
				argType := funcType.In(i)
				switch argType {
				case contextType:
					argValues[i] = reflect.ValueOf(context)
				case timerangeType:
					argValues[i] = reflect.ValueOf(context.Timerange)
				case groupsType:
					argValues[i] = reflect.ValueOf(groups)
				case stringType, scalarType, durationType, timeseriesType, valueType:
					resultI, err := evalTo(nextArgument(), argType)
					if err != nil {
						return nil, err
					}
					argValues[i] = reflect.ValueOf(resultI)
				case reflect.PtrTo(stringType), reflect.PtrTo(scalarType), reflect.PtrTo(durationType), reflect.PtrTo(timeseriesType), reflect.PtrTo(valueType):
					arg := nextArgument()
					if arg != nil {
						resultI, err := evalTo(nextArgument(), argType)
						if err != nil {
							return nil, err
						}
						// make a pointer to resultI:
						ptrValue := reflect.New(argType)
						ptrValue.Elem().Set(reflect.ValueOf(resultI))
						argValues[i] = ptrValue
					} else {
						argValues[i] = reflect.Zero(argType)
					}
				case expressionType:
					argValues[i] = reflect.ValueOf(nextArgument())
				case reflect.PtrTo(expressionType):
					arg := nextArgument()
					if arg != nil {
						ptrValue := reflect.New(argType)
						ptrValue.Elem().Set(reflect.ValueOf(arg))
						argValues[i] = ptrValue
					} else {
						argValues[i] = reflect.Zero(argType)
					}
				default:
					panic(fmt.Sprintf("Argument to MakeFunction requests invalid type %+v.", argType))
				}
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
			panic("MakeFunction built with function that doesn't return 1 or 2 things.")
		},
	}

}
