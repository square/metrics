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
	"sync"
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
		panic("MakeFunction expects a function as input.")
	}
	funcType := funcValue.Type()
	if funcType.IsVariadic() {
		panic("MakeFunction's argument cannot be variadic.")
	}
	if funcType.NumOut() == 0 {
		panic("MakeFunction's argument function must return a value.")
	}
	if funcType.NumOut() > 2 {
		panic("MakeFunction's argument function must return at most two values.")
	}
	if !funcType.Out(0).ConvertibleTo(valueType) {
		panic("MakeFunction's argument function's first return type must be convertible to `function.Value`.")
	}
	if funcType.NumOut() == 2 && !funcType.Out(1).ConvertibleTo(errorType) {
		panic("MakeFunction's argument function's second return type must convertible be `error`.")
	}

	requiredArgumentCount := 0
	optionalArgumentCount := 0
	allowsGroupBy := false
	for i := 0; i < funcType.NumIn(); i++ {
		argType := funcType.In(i)
		switch argType {
		case contextType, timerangeType:
		// Asks for part of context.
		case groupsType:
			// asks for groups
			allowsGroupBy = true
		case stringType, scalarType, durationType, timeseriesType, valueType, expressionType:
			// An ordinary argument.
			if optionalArgumentCount > 0 {
				panic("Non-optional arguments cannot occur after optional ones.")
			}
			requiredArgumentCount++
		case reflect.PtrTo(stringType), reflect.PtrTo(scalarType), reflect.PtrTo(durationType), reflect.PtrTo(timeseriesType), reflect.PtrTo(valueType), reflect.PtrTo(expressionType):
			// An optional argument
			optionalArgumentCount++
		default:
			panic(fmt.Sprintf("MetricFunction function argument asks for unsupported type: cannot supply argument %d of type %+v.", i, argType))
		}
	}
	// The function has been checked and inspected.
	// Now, generate the corresponding MetricFunction.

	return MetricFunction{
		FunctionName:  name,
		MinArguments:  requiredArgumentCount,
		MaxArguments:  requiredArgumentCount + optionalArgumentCount,
		AllowsGroupBy: allowsGroupBy,
		// Compute does a lot of reflection to get this to work.
		Compute: func(context EvaluationContext, arguments []Expression, groups Groups) (Value, error) {

			// nextArgument will extract the next argument from the expression list `arguments`.
			// if there are not more to return, it will return nil.
			expressionArgument := 0
			nextArgument := func() Expression {
				if expressionArgument >= len(arguments) {
					return nil
				}
				arg := arguments[expressionArgument]
				expressionArgument++
				return arg
			}

			// evalTo takes an expression and a reflect.Type and evaluates to the appropriate type.
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
				panic("Unreachable :: Unknown type to evaluate to")
			}

			// argumentFuncs holds functions to obtain the Value arguments.
			argumentFuncs := make([]func() (interface{}, error), funcType.NumIn())

			// provideValue takes any value, and returns a function that returns it.
			provideValue := func(x interface{}) func() (interface{}, error) {
				return func() (interface{}, error) {
					return x, nil
				}
			}

			// provideZeroValue takes a type, and returns a function that returns the zero-value for that type.
			provideZeroValue := func(t reflect.Type) func() (interface{}, error) {
				return provideValue(reflect.Zero(t).Interface())
			}

			// ptrTo takes a value and returns a pointer to that value.
			ptrTo := func(x interface{}) interface{} {
				ptr := reflect.New(reflect.TypeOf(x))
				ptr.Elem().Set(reflect.ValueOf(x))
				return ptr.Interface()
			}

			for i := range argumentFuncs {
				argType := funcType.In(i)
				switch argType {
				case contextType:
					argumentFuncs[i] = provideValue(context)
				case timerangeType:
					argumentFuncs[i] = provideValue(context.Timerange)
				case groupsType:
					argumentFuncs[i] = provideValue(groups)
				case stringType, scalarType, durationType, timeseriesType, valueType:
					arg := nextArgument()
					argumentFuncs[i] = func() (interface{}, error) {
						return evalTo(arg, argType)
					}
				case reflect.PtrTo(stringType), reflect.PtrTo(scalarType), reflect.PtrTo(durationType), reflect.PtrTo(timeseriesType), reflect.PtrTo(valueType):
					arg := nextArgument()
					if arg == nil {
						argumentFuncs[i] = provideZeroValue(argType)
					} else {
						argumentFuncs[i] = func() (interface{}, error) {
							resultI, err := evalTo(arg, argType)
							if err != nil {
								return nil, err
							}
							return ptrTo(resultI), nil
						}
					}
				case expressionType:
					argumentFuncs[i] = provideValue(nextArgument())
				case reflect.PtrTo(expressionType):
					arg := nextArgument()
					if arg == nil {
						argumentFuncs[i] = provideZeroValue(argType)
					} else {
						argumentFuncs[i] = provideValue(ptrTo(arg))
					}
				default:
					panic(fmt.Sprintf("Unreachable :: Argument to MakeFunction requests invalid type %+v.", argType))
				}
			}

			// Now we evaluate the functions in parallel.

			waiter := sync.WaitGroup{}
			argValues := make([]reflect.Value, funcType.NumIn())
			errors := make(chan error, funcType.NumIn())
			for i := range argValues {
				i := i
				waiter.Add(1)
				go func() {
					defer waiter.Done()
					arg, err := argumentFuncs[i]()
					if err != nil {
						errors <- err
						return
					}
					argValues[i] = reflect.ValueOf(arg)
				}()
			}
			waiter.Wait() // Wait for all the arguments to be evaluated.

			if len(errors) != 0 {
				return nil, <-errors
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
			panic("Unreachable :: MakeFunction built with function that doesn't return 1 or 2 things.")
		},
	}

}
