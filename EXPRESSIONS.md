
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

### `SeriesListValue`

A `SeriesListValue` is a collection of series; the `SeriesList` as a whole has a `Timerange` (start, end, interval) which applies to every series inside it.
Each series individually has a collection of tag (key, value) pairs and a sequence of sampled metric values associated to it.

They cannot be implictly converted into any other type.

### `NumberValue`

`NumberValue`s come from numeric literals. They are implictly converted to `SeriesListValue` containing a single, tagless series having constant value whenever a `SeriesList` is expected.

They cannot be implictly converted into any type other than `SeriesList`.

### `DurationValue`

`DurationValue`s come from duration literals. They cannot be implictly converted into any other type.

### `StringValue`

`StringValue`s come from string literals. They can be implictly converted only into `Duration`s.