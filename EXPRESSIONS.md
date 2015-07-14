
# Expressions

## Literals and Names

Expressions are composed of literals, operators, functions, and other syntactic constructs. Several types of names and literals are available.

### Names

Metric names and tag keys are both represented by identifiers. An identifier consists of either:

* letters, numbers, and periods (but must start with a letter)
* an arbitrary string (that contains no backticks) enclosed in backticks (`\``)

Some examples of valid identifiers:
```
cpu
host.cpu
host.cpu.median
`cpu`
`host.cpu`
`host.cpu.median`
`host-cpu`
```

### Strings

Strings represent tag values. They are enclosed in single or double quotes, with escape sequences for backticks, backslashes, and quotes.

Some examples of valid strings:
```
"Hello, world!"
'Hello, world!'
"Backslashes should be \\ escaped"
"Similarly, \" escape your quotes"
"You don't need to escape 'quotes' that differ from the enclosing ones"
"But \'you are allowed to if you want\'"
```

### Numbers

Numbers are signed sequences of digits, possibly with a decimal component and/or an exponent.

Here are examples of some valid number literals:

```
0
20
7.42
1e-5
-1E9
-2
```

### Durations

Durations are signed integers with unit suffixes that represent an interval of time.

The following units are supported:

* ms (millisecond)
* s (second)
* m (minute)
* h, hr (hour)
* d (day, 24 hours)
* w (week, 7 days)
* M, mo (month, 30 days)
* y, yr (year, 365 days)

Some examples:


Some examples:

```
10h
5m
-4s
0s
24hr
1yr
2mo
2M
-30000ms
```

## Values

Expressions and literals in MQE evaluate to one of 5 types of *values*. Value types are checked for correctness when functions are evaluated.
Only the `SeriesList` type is a legal result in a `select` query.

### `SeriesList`

A `SeriesList` is a collection of series; the `SeriesList` as a whole has a `Timerange` (start, end, interval) which applies to every series inside it.
Each series individually has a collection of tag (key, value) pairs and a sequence of sampled metric values associated to it.

They cannot be implictly converted into any other type.

### `NumberValue`

`NumberValue`s come from numeric literals. They are implictly converted to `SeriesList` containing a single, tagless series having constant value whenever a `SeriesList` is expected.

They cannot be implictly converted into any type other than `SeriesList`.

### `DurationValue`

`DurationValue`s come from duration literals. They cannot be implictly converted into any other type.

### `StringValue`

`StringValue`s come from string literals. They can be implictly converted only into `Duration`s.

## Metrics

When an identifier (name) is encountered on its own, it is interpreted as a metric-fetch expression.
MQE will fetch all series having the particular metric name that satisfy the `where` clause predicate.

### Metric Predicates

A metric fetch expression can be optionally followed by square braces enclosing a predicate which will be applied to the particular predicate.

`cpu[app = 'metrics-indexer']` will perform a fetch of `cpu` metrics only for series which have the tag `metrics-indexer`. Arbitrary predicates can be used here, so the following is also legal:

`cpu[dc = 'west' or app != 'metrics-indexer']`

## Functions

Functions are the principle way in which expressions are transformed. A function call looks like a function call in C or Java or Go. For example:

`transform.moving_average( cpu, 10m )`

### Group By

Aggregation functions take an optional `group by` clause, which takes a comma-separated list of tag keys on which to group input series.

The function `aggregate.sum` takes a series list and condenses it into a single series whose values are the sum of all values in the original series list:

`aggregate.sum( cpu )`

The `group by` clause can be used to condense groups of series together, rather than combining every series at once:

`aggregate.sum( cpu group by app )`

will produce one series for each app tracked by the `cpu` metric. All tags other than those which are `group by`-ed will be dropped.

### Pipes

A sugar is provided to make it easier to chain many functions together. The first argument to a function can be "piped" into it:

`f(x, ...) === x | f(...)`

For example, using pipe syntax, rather than writing

`transform.moving_average(cpu, 10m)`

we can instead write

`cpu | transform.moving_average(10m)`

If the function takes only one argument, then the parentheses can be ommitted when using pipes:

`cpu | aggregate.sum`

will pipe the `cpu` metric fetch into `aggregate.sum`.

Keep in mind that this is only syntactic sugar. MQE treats piped function calls and ordinary function calls in an identical manner.
In particular, this has no effect on the order of argument evaluation.

These calls can be chained together for ease of use. For example:

`cpu | aggregate.sum(group by app) | transform.derivative | transform.moving_average(2hr)`

This expression sums the cpu per-app, computes the derivative of these, and then smooths the data with a 2 hour moving average.
