MQE provides [a lot of built-in functions](https://github.com/square/metrics/wiki/Function-Reference). If MQE doesn't have what you need, you can easily register your own functions for MQE to use while it's running. This tutorial provides a basic walkthrough on creating your own custom functions to plug in to MQE.

### `identity` - the simplest example function

To start, let's just make a simple `identity` function that takes in one argument, and returns it without doing anything.

```
package example

import "github.com/square/metrics/api"
import "github.com/square/metrics/function"

var Identity function.MetricFunction = function.MakeFunction(
    "identity",
    func(input api.SeriesList) api.SeriesList {
        return input
    },
)

func init() {
    registry.MustRegister(Identity)
}
```

That was easy! `function.MakeFunction` creates a `function.MetricFunction` for us by reflecting on the provided `func` argument. `registry.MustRegister` adds the function to the default global registry, and panics if there's a problem adding it (for example, if `identity` were already defined). MQE takes care of evaluating the argument and passing it in for us.

### `divide` - asking for multiple arguments

Let's try writing a `divide` function that takes in two scalars as arguments and divides them:

```
package example

import "github.com/square/metrics/api"
import "github.com/square/metrics/function"

var Divide function.MetricFunction = function.MakeFunction(
    "divide",
    func(numerator float64, denominator float64) (float64, error) {
        if denominator == 0 {
            return 0, fmt.Errorf("cannot divide %f by 0", numerator)
        }
        return numerator / denominator, nil
    },
)

func init() {
    registry.MustRegister(Divide)
}
```

As you can see, we can return evaluation errors as well. If the returned error isn't `nil`, MQE will stop evaluation and report it.

When you build a `MetricFunction` with `MakeFunction`, MQE will automatically evaluate its arguments in parallel.

### `timerange` - using the query's timerange

Sometimes, we need a little bit of extra information to evaluate a function. For example, `transform.derivative` needs to know the resolution of the query in order to be able to scale the result appropriately.

If we ask for a `Timerange` we'll get the timerange used for the expression. Let's make a simple function that returns the query time as a string. We can use it to set a tag in the result.

```
package example

import "github.com/square/metrics/api"
import "github.com/square/metrics/function"

var Timestamp function.MetricFunction = function.MakeFunction(
    "timestamp",
    func(timerange api.Timerange) string {
        return fmt.Sprintf("%+v", timerange.End())
    },
)

func init() {
    registry.MustRegister(Timestamp)
}
```

Note that this function would be called like `timestamp()` in an MQE query; the `api.Timerange` is obtained from the MQE context, not as a formal parameter.

### `previous` - adjusting the evaluation context

Sometimes we want to be able to modify the context that is to be used in evaluation. Let's create a helper function analogous to `transform.timeshift` which shifts the timerange of the query to precisely one interval before the current query. (So, for example, `select previous(foo) from -20m to now` will be the same as `select foo from -40m to -20m`).

```
package example

import "github.com/square/metrics/api"
import "github.com/square/metrics/function"

var Previous function.MetricFunction = function.MakeFunction(
    "previous",
    func(expression function.Expression, context function.EvaluationContext) (function.Value, error) {
        timerange := context.Timerange()                       // Get the timerange
        newTimerange := timerange.Shift(-timerange.Duration()) // Shift the timerange by its duration
        newContext := context.WithTimerange(newTimerange)      // Create a new context
        return expression.Evaluate(newContext)                 // Run the expression
    },
)

func init() {
    registry.MustRegister(Previous)
}
```

If we don't want MQE to evaluate an argument for us (for example, in this case, we need to change the `context` that it evaluates in *before* it's evaluated) then we ask for a `function.Expression`. The argument will be passed to us in an unevaluated state. If we don't evaluate it, it won't be evaluated at all! This also means that we lose the benefit of automatically parallelized arguments.

In addition, we asked for a copy of the `EvaluationContext`. Once we have the adjusted timerange, we write `context.WithTimerange(newTimerange)` to obtain a copy of the context where its timerange has been replaced by the one specified.

Lastly, we call `Evaluate` with the new context. The result is a `function.Value`, which can be any of a *string*, *scalar*, *series list*, *duration*, or *tagged scalar*. `Value`s are the type that MQE uses to represent the result of an `Expression`.

### `negate` - operating on time series data

When you're using MQE, most of your functions will probably operate on time-series data. Let's see how we can make a `negate` function that multiplies every value by -1.

```
package example

import "github.com/square/metrics/api"
import "github.com/square/metrics/function"

var Negate function.MetricFunction = function.MakeFunction(
    "negate",
    func(list api.SeriesList) api.SeriesList {
        result := api.SeriesList{
            Series: make([]api.Timeseries, len(list.Series))
        }
        for i, line := range list.Series {
            result.Series[i] = api.Timeseries{
                Values: make([]float64, len(line.Values)),
                TagSet: line.TagSet,
            }
            for j := range line.Values {
                result.Series[i].Values[j] = -line[j]
            }
        }
        return result
    },
)

func init() {
    registry.MustRegister(Negate)
}
```

Every time series has an associated `api.TagSet`. Note that it's a type definition of `map[string]string`; but you should treat them as *immutable* once you've created them. In particular, modifying tagsets obtained from your functions arguments can have very unexpected results, as they may be shared between different values in your computation.

