
# Expressions

## Values

Expressions in MQE evaluate to one of 5 types of values. Some of these values can be implictly converted from one to another.
Only the first type, the `SeriesList`, is a legal top-level result in a `select` query.

### `SeriesList`

A `SeriesList` has a `Name`, a set of `Series`, and a `Timerange`.

#### `Name`

The name identifies the `SeriesList`. As an expression is evaluated, its name is updated to describe the transformations and operations which have been applied to it.
Names are initially the name of the metric which was fetched.

#### `Timerange`

The `Timerange` has a start, end, and resolution. The resolution measures the time between samples.
Both the start and end times are inclusive. Together with resolution, they uniquely determine the number of samples in each timeseries.
Missing data will be filled with `NaN` in order to ensure that their values slices are all the correct length.

#### `Series`

Each `Series` has a sequence of `Values`, which are the measured metric data. In addition, it has a `TagSet`, which is a map of `string => string` that describes the series in the `SeriesList`.
In general, all `Series` in a given `SeriesList` should use the same tag keys, and have distinct tag values so that they can be differentiated.

### Number

Numbers are valid values. They can be implictly converted to a Timeseries of constant value. In general, numerical operations will convert a number to a `SeriesList` before operating on them.

### Duration

Durations are values which correspond to a length of time. When written as a literal, they are suffixed with a unit. The following units are supported:

* ms (millisecond)
* s (second)
* m (minute)
* h, hr (hour)
* d (day, 24 hours)
* w (week, 7 days)
* M, mo (month, 30 days)
* y, yr (year, 365 days)

Durations and numbers are distinct types of data, and it is not possible to implictly convert between them. Durations cannot be converted into `SeriesList`s.

### String

Strings are string data which generally represent key values or metric names. They are written as literals enclosed in single or double quotes.