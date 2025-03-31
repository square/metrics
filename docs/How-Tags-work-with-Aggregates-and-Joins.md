MQE supports aggregations and joins. While generally these will "do what you want them to", occasionally you might run into some trouble if you don't understand how they work. This page provides a detailed guide into how these operations are implemented and defined.

### Terminology

MQE treats "series lists" as a basic type of value on which it operates. A series list is a list of time series lines, where each line has an associated tag set. In general, it's assumed that tag sets uniquely identify each time series line, but MQE can handle the presence of duplicates.

## Aggregation Functions

Aggregation functions take a series list as an argument containing many individual time series lines, and combine these series into fewer time series lines.
The following aggregation functions are supported:

* `aggregate.sum`
* `aggregate.mean`
* `aggregate.min`
* `aggregate.max`

Aggregation functions take only one value argument: the list of series to aggregate on. They will collapse all series in the argument list into a single resulting series.
For example, given a series list produced by `latency`:

| (series†) | metric name | tags                         | series values |
|:---------:|:-----------:|:----------------------------:|:-------------:|
| (A1)      | latency     | app: ui,     env: staging    | `1 2 1`       |
| (A2)      | latency     | app: ui,     env: production | `3 3 3`       |
| (A3)      | latency     | app: server, env: staging    | `0 0 1`       |
| (A4)      | latency     | app: server, env: production | `2 2 0`       |

> † The series identifiers are purely illustrative - internally, individual timeseries are identified solely by their metric name and tags.

Querying `aggregate.sum( latency )` computes the following result, a list of a single series:

| (series†)         | metric name            | tags      | series values |
|:-----------------:|:----------------------:|:---------:|:-------------:|
|(A1)+(A2)+(A3)+(A4)| aggregate.sum(latency) | (no tags) | `6 7 5`       |

Aggregators in general treat missing data points (internally represented as `NaN`) as though they were not present. For example, consider the following:

| (series†) | metric name | tags                         | series values |
|:---------:|:-----------:|:----------------------------:|:-------------:|
| (A1)      | latency     | app: ui,     env: staging    | `8   NaN 2  ` |
| (A2)      | latency     | app: ui,     env: production | `8   6   NaN` |
| (A3)      | latency     | app: server, env: staging    | `NaN 9   NaN` |
| (A4)      | latency     | app: server, env: production | `8   3   8  ` |

The result of `aggregate.mean` is:

| (series)          | metric name            | tags      | series values |
|:-----------------:|:----------------------:|:---------:|:-------------:|
|(A1)+(A2)+(A3)+(A4)| aggregate.sum(latency) | (no tags) | `8 6 5`       |

Occasionally, we don't want to combine all time series lines into a single line, but rather into several based on their tags. We can *group* the result by a tag (or set of tags), preserving these tags in the resulting time series lines.

Consider the metric `latency` again from above:

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

You can group by multiple tags by listing them, separated by commas, like this: `aggregate.sum(latency group by app, env)`.

## Joins

Aggregates are used to condense a large collection of time series lines into a smaller number. Joins take two distinct series list values and create a new list using time series lines from each.

MQE supports 4 binary infix numerical operators, implemented as joins:

* `+`
* `-`
* `*`
* `/`

Both arguments to the operator are each evaluated to a list of time series lines.

A *join* is performed between the two lists:

* every line from the left list is paired with every line from the right list
* pairs are evaluated on their tags: if two lines in a pair both have some tag `K`, and their values `VL` and `VR` are different, the pair is discarded
* all remaining pairs of lines are combined into individual lines, with tag sets formed from the union of the pair

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

There are 4 candidate result pairs: (A1 + B1), (A1 + B2), (A2 + B1), and (A2 + B2).

Then, we evaluate them for conflicts.

* `(A1 + B1)` is not conflicting `{ app: ui, env: staging, method: rpc }`
* `(A1 + B2)` conflicts on `app`: `ui` and `server`
* `(A2 + B1)` is not conflicting `{ app: ui, host: h0, method: rpc }`
* `(A2 + B2)` conflicts on `app`: `ui` and `server`



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

