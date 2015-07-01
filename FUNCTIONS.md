
# Functions

The metrics query engine supports functions to operate on metrics data.
Built in functions can be classified broadly into operators, aggregators, transformers, and filters.

## Operators

MQE supports 4 binary infix numerical operators:

* `+`
* `-`
* `*`
* `/`

The left- and right-hand sides of the operator must evaluate to series lists (or be convertible to series lists).
Once they are evaluated, the [inner join](https://en.wikipedia.org/wiki/Join_(SQL\)#Inner_join) of the left and the right lists are computed.
Each series from the left is paired with each series from the right, allowing those pairs which have no conflicting tags.

For example, in the query `MetricA + MetricB`, if our series lists look like this:

|MetricA TagSet               |MetricA Values|MetricB TagSet              |MetricB Values|
|:---------------------------:|:------------:|:--------------------------:|:-------------:
| app: ui, env: staging       | 1 2 1        |app: ui, latency: max       | 0 0 4        |
| app: ui, env: production    | 3 3 3        |app: ui, latency: median    | 1 1 1        |
| app: server, env: staging   | 0 0 1        |app: server, latency: max   | 2 1 0        |
| app: server, env: production| 2 2 0        |app: server, latency: median| 3 2 3        |

Then the result of the operation will be:

|`MetricA + MetricB` TagSet                   |`MetricA + MetricB` Values|
|:-------------------------------------------:|:------------------------:|
|app: ui, env: staging, latency: max          | 1 2 5                    |
|app: ui, env: staging, latency: median       | 2 3 2                    |
|app: ui, env: production, latency: max       | 3 3 7                    |
|app: ui, env: production, latency: median    | 4 4 4                    |
|app: server, env: staging, latency: max      | 2 1 1                    |
|app: server, env: staging, latency: median   | 3 2 4                    |
|app: server, env: production, latency: max   | 4 3 0                    |
|app: server, env: production, latency: median| 5 4 4                    |

## Aggregation Functions

Aggregation functions take a serieslist containing many individual series, and combine these series into a smaller number.
The following aggregation functions are supported:

* `aggregate.sum`
* `aggregate.mean`
* `aggregate.min`
* `aggregate.max`

Aggregation functions take only one argument: the series list to aggregate on. They will collapse all series in the argument list into a single resulting series.
For example, given a series list produced by `MetricA`, writing `aggregate.sum( MetricA )` computes the following:

|MetricA TagSet               |MetricA Values|
|:---------------------------:|:------------:|
| app: ui, env: staging       | 1 2 1        |
| app: ui, env: production    | 3 3 3        |
| app: server, env: staging   | 0 0 1        |
| app: server, env: production| 2 2 0        |

|`aggregate.sum(MetricA)` Tagset|`aggregate.sum(MetricA)` Values|
|:---------------------------:|:-------------------------------:|
| (no tags)                   | 6 7 5                           |

Aggregators treat missing data (`NaN`) as though it were not present. For example, consider the following:

|MetricB TagSet               |MetricB Values|
|:---------------------------:|:------------:|
| app: ui, env: staging       | 8   NaN 2    |
| app: ui, env: production    | 8   6   NaN  |
| app: server, env: staging   | NaN 9   NaN  |
| app: server, env: production| 8   3   8    |

The result of `aggregate.mean` is:

|`aggregate.mean(MetricB)` TagSet |`aggregate.mean(MetricB)` Values|
|:-------------------------------:|:------------------------------:|
| (no tags)                       | 8 6 5                          |

Aggregations can be grouped by individual tags. The series in the resulting series list preserve those tags which their group used.

|MetricA TagSet               |MetricA Values|
|:---------------------------:|:------------:|
| app: ui, env: staging       | 1 2 1        |
| app: ui, env: production    | 3 3 3        |
| app: server, env: staging   | 0 0 1        |
| app: server, env: production| 2 2 0        |

|`aggregate.sum(MetricA group by app)` TagSet|`aggregate.sum(MetricA group by app)` Values|
|:------------------------------------------:|:------------------------------------------:|
| app: ui                                    | 4 5 4                                      |
| app: server                                | 2 2 1                                      |

## Transformation Functions

Transformation functions modify each series in a given serieslist independently. The following transformation functions are supported:

* `transform.derivative`
* `transform.integral`
* `transform.rate`
* `transform.cumulative`
* `transform.default`
* `transform.abs`
* `transform.nan_keep_last`

### `transform.derivative(list)`

This function computes a numerical derivative for a given timeseries. Each value will be updates to be `y[i] = (x[i] - x[i-1]) / interval` where `interval` is the time between samples, measured in seconds.
The first value is assigned 0.

### `transform.integral(list)`

This function computes a numerical integral for a given timeseries. Each value will be the sum of the values up to it (including itself), a left Riemann integral.
This quantity will be scaled, interpreting each value in the series as having units of `events / second`, and producing something with the units of `events`.
If you do not want this behavior, use `transform.cumulative` instead, which does not have scaling behavior.

`NaN` values are treated as 0.

### `transform.rate(list)`

This function computes the numerical derivative of a given timeseries, as in `transform.derivative`, but it bounds the result to be at least 0.
This is useful for counting timeseries. The resulting units are in `events / second` (so scaling will occur depending on the resolution of the data).

### `transform.cumulative(list)`

This function computes the raw, cumulsative sum of the values in each timeseries. It performs no scaling. `NaN` values are treated as 0.

### `transform.default(list, value)`

This function takes an extra number parameter. Any occurrence of `NaN` in any series in the list will be replaced by `value`.

### `transform.abs(list)`

This function computes the absolute value of all values in each series of the given list.

### `transform.nan_keep_last(list)`

This function replaces any `NaN` value with the last non-`NaN` value before it. For data which is very sparse, this can make graphs more readable.
Initial `NaN`s are left alone. If these need to be eliminated also, consider using `transform.default` on the result.

### `transform.timeshift(list, duration)`

This function shifts time forward or backward while computing `list`. In particular, `duration` is evaluated first, and then `list` (an arbitrary expression) is computed with a shifted timerange.
As a result, the computed result will contain data fetched before or after the rest of the expression.

A positive duration is forward in time, while a negative duration is backwards in time.

### `transform.moving_average(list, duration)`

This function computes a moving average over the given duration for `list`.
The function automatically extends the timerange of `list` in order to achieve a genuine moving average outside of the timerange provided by list.

Each value is replaced by the average of all samples (including itself) in the interval of length `duration` prior to itself. `NaN` values are treated as absent.

### `tramsform.alias(list, name)`

This function renames the given list to be called by the given name.

## Filter Functions