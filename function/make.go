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

// An OptionName identifies an option passed to MakeFunction
type OptionName int

const (
	InvalidOption OptionName = iota // InvalidOption represents an invalid option
	WidenBy                         // WidenBy indicates that the given duration Argument index, or the number of Slots should be used to extend the timerange in the query
	ShiftBy                         // ShiftBy acts as WidenBy, but with opposite sign
)

// String makes the option name human-readable.
func (n OptionName) String() string {
	switch n {
	case WidenBy:
		return "WidenBy"
	case ShiftBy:
		return "ShiftBy"
	default:
		return "Invalid"
	}
}

// An Option is an annotation used to add extra functionality to the generated MetricFunction.
type Option struct {
	Name  OptionName  // Name identifies the option type
	Value interface{} // Value is the value associated with the option
}

// Argument represents an argument index
type Argument int

// Slot represents a number of slots
type Slot int

// MakeFunction is a convenient way to use type-safe functions to
// construct MetricFunctions without manually checking parameters.
func MakeFunction(name string, function interface{}, options ...Option) MetricFunction {
	funcValue := reflect.ValueOf(function)
	if funcValue.Kind() != reflect.Func {
		panic(fmt.Sprintf("MakeFunction for function `%s` expects a function as second parameter.", name))
	}
	funcType := funcValue.Type()
	if funcType.IsVariadic() {
		panic(fmt.Sprintf("MakeFunction for function `%s`'s arguments cannot be variadic.", name))
	}
	if funcType.NumOut() == 0 {
		panic(fmt.Sprintf("MakeFunction for function `%s`'s argument function must return a value.", name))
	}
	if funcType.NumOut() > 2 {
		panic(fmt.Sprintf("MakeFunction for function `%s`'s argument function must return at most two values.", name))
	}
	if !funcType.Out(0).ConvertibleTo(valueType) && funcType.Out(0) != timeseriesType {
		panic(fmt.Sprintf("MakeFunction for function `%s`'s first return type must be convertible to `function.Value`.", name))
	}
	if funcType.NumOut() == 2 && !funcType.Out(1).ConvertibleTo(errorType) {
		panic(fmt.Sprintf("MakeFunction for function `%s`'s second return type must convertible be `error`.", name))
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
		case stringType, scalarType, scalarSetType, durationType, timeseriesType, valueType, expressionType:
			// An ordinary argument.
			if optionalArgumentCount > 0 {
				panic(fmt.Sprintf("MakeFunction for function `%s` has non-optional arguments after optional ones.", name))
			}
			requiredArgumentCount++
		case reflect.PtrTo(stringType), reflect.PtrTo(scalarType), reflect.PtrTo(scalarSetType), reflect.PtrTo(durationType), reflect.PtrTo(timeseriesType), reflect.PtrTo(valueType), reflect.PtrTo(expressionType):
			// An optional argument
			optionalArgumentCount++
		default:
			panic(fmt.Sprintf("MakeFunction for function `%s` function argument asks for unsupported type: cannot supply argument %d of type %+v.", name, i, argType))
		}
	}
	// The function has been checked and inspected.
	// Now, generate the corresponding MetricFunction.

	resultFunction := MetricFunction{
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
			// If an Expression is requested, it just returns it.
			evalTo := func(expression Expression, resultType reflect.Type) (interface{}, error) {
				switch resultType {
				case expressionType:
					return expression, nil
				case stringType:
					return EvaluateToString(expression, context)
				case scalarType:
					return EvaluateToScalar(expression, context)
				case scalarSetType:
					return EvaluateToScalarSet(expression, context)
				case durationType:
					return EvaluateToDuration(expression, context)
				case timeseriesType:
					return EvaluateToSeriesList(expression, context)
				case valueType:
					return expression.Evaluate(context)
				}
				panic(fmt.Sprintf("Unreachable :: Attempting to evaluate to unknown type %+v", resultType))
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
					argumentFuncs[i] = provideValue(context.Timerange())
				case groupsType:
					argumentFuncs[i] = provideValue(groups)
				case stringType, scalarType, scalarSetType, durationType, timeseriesType, valueType, expressionType:
					arg := nextArgument()
					argumentFuncs[i] = func() (interface{}, error) {
						return evalTo(arg, argType)
					}
				case reflect.PtrTo(stringType), reflect.PtrTo(scalarType), reflect.PtrTo(scalarSetType), reflect.PtrTo(durationType), reflect.PtrTo(timeseriesType), reflect.PtrTo(valueType), reflect.PtrTo(expressionType):
					arg := nextArgument()
					if arg == nil {
						argumentFuncs[i] = provideZeroValue(argType)
					} else {
						argumentFuncs[i] = func() (interface{}, error) {
							resultI, err := evalTo(arg, argType.Elem())
							if err != nil {
								return nil, err
							}
							return ptrTo(resultI), nil
						}
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

			if len(output) == 2 && output[1].Interface() != nil {
				return nil, output[1].Interface().(error)
			}
			switch funcType.Out(0) {
			case stringType:
				return StringValue(output[0].Interface().(string)), nil
			case scalarType:
				return ScalarValue(output[0].Interface().(float64)), nil
			case scalarSetType:
				return output[0].Interface().(ScalarSet), nil
			case durationType:
				return DurationValue{"", output[0].Interface().(time.Duration)}, nil
			case timeseriesType:
				return SeriesListValue(output[0].Interface().(api.SeriesList)), nil
			default:
				return output[0].Interface().(Value), nil
			}
		},
	}

	setFlags := map[string]bool{}

	for _, option := range options {
		switch option.Name {
		case WidenBy, ShiftBy:
			if setFlags["Widen"] {
				panic(fmt.Sprintf("MakeFunction for function `%s` given option %s with parameter %v; but already had a WidenBy/ShiftBy Option", name, option.Name, option.Value))
			}
			setFlags["Widen"] = true
			sign := -1
			if option.Name == ShiftBy {
				sign = 1
			}
			switch value := option.Value.(type) {
			case Argument:
				resultFunction.Widen = func(widen WidestMode, arguments []Expression) time.Time {
					result := widen.Current
					if int(value) >= len(arguments) {
						return result
					}
					literalInterface, ok := arguments[int(value)].(LiteralExpression)
					if !ok {
						return result
					}
					literalValue := literalInterface.Literal()
					if literalValue == nil {
						return result
					}
					duration, ok := literalValue.(time.Duration)
					if !ok {
						return result
					}
					widen.AddTime(widen.Current.Add(time.Duration(sign) * duration))
					if option.Name == ShiftBy {
						return result.Add(duration)
					}
					return result
				}
			case Slot:
				resultFunction.Widen = func(widen WidestMode, arguments []Expression) time.Time {
					widen.AddTime(widen.Current.Add(-widen.Resolution))
					return widen.Current
				}
			default:
				panic(fmt.Sprintf("MakeFunction for function `%s` given option %s with value %v of unsupported type %T; must be either function.Argument or function.Slot", name, option.Name, option.Value, option.Value))
			}
		default:
			panic(fmt.Sprintf("MakeFunction for function `%s` given unrecognized option %s (with argument %v)", name, option.Name, option.Value))
		}
	}

	return resultFunction
}

var stringType = reflect.TypeOf("")
var scalarType = reflect.TypeOf(float64(0.0))
var scalarSetType = reflect.TypeOf(ScalarSet{})
var durationType = reflect.TypeOf(time.Duration(0))
var timeseriesType = reflect.TypeOf(api.SeriesList{})
var valueType = reflect.TypeOf((*Value)(nil)).Elem()
var expressionType = reflect.TypeOf((*Expression)(nil)).Elem()
var groupsType = reflect.TypeOf(Groups{})
var contextType = reflect.TypeOf(EvaluationContext{})
var timerangeType = reflect.TypeOf(api.Timerange{})

var errorType = reflect.TypeOf((*error)(nil)).Elem()
