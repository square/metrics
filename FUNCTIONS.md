
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

* aggregate.sum
* aggregate.mean
* aggregate.min
* aggregate.max

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
| (no tags)                       | 9 6 5                          |

Aggregations can be grouped by individual tags. The series in the resulting series list preserve those tags which their group used.

|MetricA TagSet               |MetricA Values|
|:---------------------------:|:------------:|
| app: ui, env: staging       | 1 2 1        |
| app: ui, env: production    | 3 3 3        |
| app: server, env: staging   | 0 0 1        |
| app: server, env: production| 2 2 0        |

|`aggregate.sum(MetricA group by app)` TagSet|`aggregate.sum(MetricA group by app)` Values|
| app: ui                                    | 4 5 4                                      |
| app: server                                | 2 2 1                                      |

## Transformation Functions

## Filter Functions