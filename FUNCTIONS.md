
# Functions

The metrics query engine supports functions to operate on metrics data.
Built in functions can be classified broadly into operators, aggregators, transformers, and filters.

## Operators

MQE supports 4 binary infix numerical operators:

* `+`
* `-`
* `*`
* `/`

Both sides of the operator are first evaluated to lists of timeseries.
An inner join is performed between the two lists;
every timeseries on the left is paired with every timeseries on the right, discarding those pairs which have conflicting tagsets.

Two tagsets are "conflicting" if they both possess some tag key, but disagree on its value.

The resulting pairs are then combined using the arithmetic operator, in a pairwise manner on the timeseries data.
The tagsets of the resulting timeseries are the unions of the tagsets of the members of the original pairs.

For example, imagine performing the query:

`latency.method + latency.connection`

We evaluate `latency.method` and `latency.connection` and find the following timeseries lists are the results:

| (series†) | metric name    | tags                    | series values |
|:---------:|:--------------:|:-----------------------:|:-------------:|
| (A1)      | latency.method | [app: ui, env: staging] | `1 1 1`       |
| (A2)      | latency.method | [app: ui, host: h0    ] | `2 2 2`       |

We evaluate `latency.connection` and find that the result is the following list of two timeseries:

| (series†) | metric name        | tags                       | series values |
|:---------:|:------------------:|:--------------------------:|:-------------:|
| (B1)      | latency.connection | [app: ui,     method: rpc] | `3 3 3`       |
| (B2)      | latency.connection | [app: server, method: rpc] | `4 4 4`       |

† The series identifiers are purely illustrative - internally, individual timeseries are identified solely by their metric name and tags.

There are 4 candidate result pairs: (A1 + B1), (A1 + B2), (A2 + B1), and (A2 + B2).

Of these, (A1 + B2) and (A2 + B2) are both conflicting, since (A1) and (A2) have `app: ui` while (B2) has `app: server`.
(A1 + B1) and (A2 + B1) are not conflicting, however.

Therefore the following result is reached:

| (series†) | metric name                         | tags                                               | series values |
|:---------:|:-----------------------------------:|:--------------------------------------------------:|:-------------:|
| (A1)+(B1) | latency.method + latency.connection | [app: ui,     env: staging,           method: rpc] | `4 4 4`       |
| (A2)+(B1) | latency.method + latency.connection | [app: server,               host: h0, method: rpc] | `5 5 5`       |

Important things to note:

* Not every timeseries from the argument series lists will be represented in the output. (B2) conflicted with every tag from `latency.method` and therefore isn't represented in the result.
* Timeseries can be matched with more than one series in the result. (B1) matched with *both* (A1) and (A2)
* Series from one size of the operator aren't joined with one another. For example, although (A1) and (A2) are not conflicting, they are both from the left argument, so (A1)+(A2) isn't in the result.
* If series have no tag keys in common, they will *always* match.
* The left and right sides of the operator may contain different numbers of timeseries.

A larger example is provided here:

Consider the query `latency.method + latency.connection` over the following data:

| (series†) | metric name        | tags                         | series values |
|:---------:|:------------------:|:----------------------------:|:-------------:|
| (A1)      | latency.method     | app: ui,     env: staging    | `1 2 1`       |
| (A2)      | latency.method     | app: ui,     env: production | `3 3 3`       |
| (A3)      | latency.method     | app: server, env: production | `0 0 1`       |
|           |                    |                              |               |
| (B1)      | latency.connection | app: ui,     method: rpc     | `0 0 4`       |
| (B2)      | latency.connection | app: ui,     method: http    | `1 1 1`       |
| (B3)      | latency.connection | app: server, method: rpc     | `2 1 0`       |
| (B4)      | latency.connection | app: server, method: http    | `3 2 3`       |

The query results in a list of 6 timeseries:

| (series†) | metric name                         | tags                                       | series values |
|:---------:|:-----------------------------------:|:------------------------------------------:|:-------------:|
| (A1)+(B1) | latency.method + latency.connection | app: ui,     env: staging,    method: rpc  | `1 2 5`       |
| (A1)+(B2) | latency.method + latency.connection | app: ui,     env: staging,    method: http | `2 3 2`       |
| (A2)+(B1) | latency.method + latency.connection | app: ui,     env: production, method: rpc  | `3 3 7`       |
| (A2)+(B2) | latency.method + latency.connection | app: ui,     env: production, method: http | `4 4 4`       |
| (A3)+(B3) | latency.method + latency.connection | app: server, env: production, method: rpc  | `2 1 1`       |
| (A3)+(B4) | latency.method + latency.connection | app: server, env: production, method: http | `3 2 4`       |

## Aggregation Functions

Aggregation functions take a serieslist containing many individual series, and combine these series into a smaller number.
The following aggregation functions are supported:

* `aggregate.sum`
* `aggregate.mean`
* `aggregate.min`
* `aggregate.max`

Aggregation functions take only one argument: the list of series to aggregate on. They will collapse all series in the argument list into a single resulting series.
For example, given a series list produced by `latency`:

| (series†) | metric name | tags                         | series values |
|:---------:|:-----------:|:----------------------------:|:-------------:|
| (A1)      | latency     | app: ui,     env: staging    | `1 2 1`       |
| (A2)      | latency     | app: ui,     env: production | `3 3 3`       |
| (A3)      | latency     | app: server, env: staging    | `0 0 1`       |
| (A4)      | latency     | app: server, env: production | `2 2 0`       |

Querying `aggregate.sum( latency )` computes the following result, a list of a single series:

| (series†)         | metric name            | tags      | series values |
|:-----------------:|:----------------------:|:---------:|:-------------:|
|(A1)+(A2)+(A3)+(A4)| aggregate.sum(latency) | (no tags) | `6 7 5`       |

Aggregators in general treat missing data (internally represented as `NaN`) as though it were not present. For example, consider the following:

| (series†) | metric name | tags                         | series values |
|:---------:|:-----------:|:----------------------------:|:-------------:|
| (A1)      | latency     | app: ui,     env: staging    | `8   NaN 2  ` |
| (A2)      | latency     | app: ui,     env: production | `8   6   NaN` |
| (A3)      | latency     | app: server, env: staging    | `NaN 9   NaN` |
| (A4)      | latency     | app: server, env: production | `8   3   8  ` |

The result of `aggregate.mean` is:

| (series†)         | metric name            | tags      | series values |
|:-----------------:|:----------------------:|:---------:|:-------------:|
|(A1)+(A2)+(A3)+(A4)| aggregate.sum(latency) | (no tags) | `8 6 5`       |

Aggregations can be grouped by individual tags. The series in the resulting series list preserve those tags which their group used.

Consider the metric `latency`:

| (series†) | metric name | tags                         | series values |
|:---------:|:-----------:|:----------------------------:|:-------------:|
| (A1)      | latency     | app: ui,     env: staging    | `1 2 1`       |
| (A2)      | latency     | app: ui,     env: production | `3 3 3`       |
| (A3)      | latency     | app: server, env: staging    | `0 0 1`       |
| (A4)      | latency     | app: server, env: production | `2 2 0`       |

If we want to find the total latency per-app then we can run `aggregate.sum(latency group by app)`. We get the following result:

| (series†) | metric name                         | tags        | series values |
|:---------:|:-----------------------------------:|:-----------:|:-------------:|
|(A1)+(A2)  | aggregate.sum(latency group by app) | app: ui     | `4 5 4`       |
|(A3)+(A4)  | aggregate.sum(latency group by app) | app: server | `2 2 1`       |

Note that only those tags which you list in the `group by` clause are preserved- even if all series in the group agree on a particular tag value, the tag will be omitted in the result.

## Transformation Functions

Transformation functions modify each series in a given serieslist independently. The following transformation functions are supported:

* `transform.derivative`
* `transform.integral`
* `transform.rate`
* `transform.cumulative`
* `transform.nan_fill`
* `transform.abs`
* `transform.nan_keep_last`

##### `transform.derivative(list)`

This function computes a numerical derivative for a given timeseries. Each value will be updates to be `y[i] = (x[i] - x[i-1]) / interval` where `interval` is the time between samples, measured in seconds.
The first value is assigned 0.

Consider performing the query `transform.derivative( disk_usage )` with a resolution of 10 seconds (30 seconds is the default resolution), where `disk_usage` is defined like this:

| (series†) | metric name | tags        | series values |
|:---------:|:-----------:|:-----------:|:-------------:|
| (A1)      | disk_usage  | app: ui     | `300 310 315` |
| (A2)      | disk_usage  | app: server | `440 435 450` |

The query `transform.deriative( disk_usage )` results in:

| (series†) | metric name                      | tags        | series values |
|:---------:|:--------------------------------:|:-----------:|:-------------:|
| (A1)      | transform.derivative(disk_usage) | app: ui     | `0  1   0.5`  |
| (A2)      | transform.derivative(disk_usage) | app: server | `0 -0.5 1.5`  |

##### `transform.integral(list)`

This function computes a numerical integral for a given timeseries. Each value will be the sum of the values up to it (including itself).
This quantity will be scaled, interpreting each value in the series as having units of `events / second`, and producing something with the units of `events`.
If you do not want this behavior, use `transform.cumulative` instead, which does not have scaling behavior.

`NaN` values are treated as 0.

Consider the query `transform.integral( request_rate )` with a resolution of 10 seconds (30 seconds is the default resolution) on the metric `request_rate` detailed below:

| (series†) | metric name  | tags        | series values |
|:---------:|:------------:|:-----------:|:-------------:|
| (A1)      | request_rate | app: ui     | `30  25  20 ` |
| (A2)      | request_rate | app: server | `120 134 150` |

The query `transform.integral( request_rate )` results in:

| (series†) | metric name                      | tags        | series values    |
|:---------:|:--------------------------------:|:-----------:|:----------------:|
| (A1)      | transform.integral(request_rate) | app: ui     | `300  550  750`  |
| (A2)      | transform.integral(request_rate) | app: server | `1200 2540 4040` |

##### `transform.rate(list)`

This function computes the numerical derivative of a given timeseries, as in `transform.derivative`, but it bounds the result to be at least 0.
This transformation is most useful on counters. The resulting units are in `events / second` (so scaling will occur depending on the resolution of the data).

Consider performing the query `transform.rate( disk_usage )` with a resolution of 10 seconds (30 seconds is the default resolution), where `disk_usage` is defined like this:

| (series†) | metric name | tags        | series values   |
|:---------:|:-----------:|:-----------:|:---------------:|
| (A1)      | disk_usage  | app: ui     | `300  310  315` |
| (A2)      | disk_usage  | app: server | `440  435  450` |

The query `transform.rate( disk_usage )` results in:

| (series†) | metric name                | tags        | series values |
|:---------:|:--------------------------:|:-----------:|:-------------:|
| (A1)      | transform.rate(disk_usage) | app: ui     | `0  1  0.5`   |
| (A2)      | transform.rate(disk_usage) | app: server | `0  0  1.5`   |

##### `transform.cumulative(list)`

This function computes the raw, cumulsative sum of the values in each timeseries. It performs no scaling. `NaN` values are treated as 0.

If scaling is desired, use `transform.integral(list)`.

Consider the query `transform.cumulative( request_counter )` with a resolution of 10 seconds (30 seconds is the default resolution) on the metric `request_rate` detailed below:

| (series†) | metric name     | tags        | series values |
|:---------:|:---------------:|:-----------:|:-------------:|
| (A1)      | request_counter | app: ui     | `30  25  20 ` |
| (A2)      | request_counter | app: server | `120 134 150` |

The query `transform.cumulative( request_counter )` results in:

| (series†) | metric name                           | tags        | series values  |
|:---------:|:-------------------------------------:|:-----------:|:--------------:|
| (A1)      | transform.cumulative(request_counter) | app: ui     | `30  55  75 `  |
| (A2)      | transform.cumulative(request_counter) | app: server | `120 254 404`  |

##### `transform.nan_fill(list, value)`

This function takes an extra number parameter. Any occurrence of `NaN` in any series in the list will be replaced by `value`.

Consider the query `transform.nan_fill(latency, 1000)`. If `latency` is detailed as below:

| (series†) | metric name | tags        | series values |
|:---------:|:-----------:|:-----------:|:-------------:|
| (A1)      | latency     | app: ui     | `80  24   15 `|
| (A2)      | latency     | app: server | `70  NaN  NaN`|

Then `transform.nan_fill(latency, 1000)` produces this result:

| (series†) | metric name                 | tags        | series values  |
|:---------:|:---------------------------:|:-----------:|:--------------:|
| (A1)      | transform.nan_fill(latency) | app: ui     | `80 24   15  ` |
| (A2)      | transform.defailt(latency)  | app: server | `70 1000 1000` |

##### `transform.abs(list)`

This function computes the absolute value of all values in each series of the given list.

Consider the query: `transform.abs(offset)` where `offset` is detailed as below:

| (series†) | metric name | tags        | series values |
|:---------:|:-----------:|:-----------:|:-------------:|
| (A1)      | offset      | app: ui     | `30  0  -15`  | 
| (A2)      | offset      | app: server | `7   6  -3 `  |

Then the query `transform.abs(offset)` results in:

| (series†) | metric name           | tags        | series values |
|:---------:|:---------------------:|:-----------:|:-------------:|
| (A1)      | transform.abs(offset) | app: ui     | `30  0  15`   | 
| (A2)      | transform.abs(offset) | app: server | `7   6  3 `   |

##### `transform.nan_keep_last(list)`

This function replaces any `NaN` value with the last non-`NaN` value before it. For data which is very sparse, this can make graphs more readable.
Initial `NaN`s are left alone. If these need to be eliminated also, consider using `transform.nan_fill` on the result.

Consider the metric `responses`:

| (series†) | metric name | tags        | series values                     |
|:---------:|:-----------:|:-----------:|:---------------------------------:|
| (A1)      | responses   | app: ui     | `NaN  NaN  3    NaN  7  Nan  NaN` | 
| (A2)      | responses   | app: server | `2    3    NaN  5    3  NaN  1  ` |

Then the query `transform.nan_keep_last(responses)` produces the result:

| (series†) | metric name                        | tags        | series values       |
|:---------:|:----------------------------------:|:-----------:|:-------------------:|
| (A1)      | transform.nan_keep_last(responses) | app: ui     | `NaN NaN 3 3 7 7 7` | 
| (A2)      | transform.nan_keep_last(responses) | app: server | `2   3   3 5 3 3 1` |

##### `transform.timeshift(list, offsetDuration)`

This function shifts time forward by the specified `offsetDuration` while computing `list`. For example, the query:

```
select transform.timeshift(metric, -1w) from -1d to now
```

would fetch metrics from one week ago. However, their resulting timestamps would be those of the past day. For example, if we want to compare this week's cpu usage with last week's, then we could write:

```
select cpu - transform.timeshift(cpu, -1w) from -1w to now
```

Keep in mind that a positive shift will go forward in time, and any data fetched from a time later than `now` will be missing.

##### `transform.moving_average(list, duration)`

This function computes a moving average over the given duration for `list`.
The function automatically extends the timerange of `list` in order to achieve a genuine moving average outside of the timerange provided by list.

Each value is replaced by the average of all samples (including itself) in the interval of length `duration` prior to itself. `NaN` values are treated as absent.

##### `transform.alias(list, name)`

This function renames the given list to be called by the given name. It does not change the series data values themselves.

## Filter Functions

Filter functions limit the number of timeseries returned by a query. Series can be sorted by their `max`, `mean`, or `min`, and ordered by `lowest` or `highest`. For example:

```
select filter.highest_mean( cpu , 5 ) from -1d to now
```

will select only the 5 series whose means are highest. The complete list of functions are:

* `filter.highest_max(list, count)`
* `filter.highest_mean(list, count)`
* `filter.highest_min(list, count)`
* `filter.lowest_max(list, count)`
* `filter.lowest_mean(list, count)`
* `filter.lowest_min(list, count)`

The value for `count` will be rounded to the nearest whole number. If, after rounding, its value is negative, the query engine will produce an error.
If the rounded `count` exceeds the number of series returned by the `list`, then all series will be retained.
