# Tutorial

This tutorial presents a walkthrough of MQE's syntax and features. 

# Describes

If we want to know all of the metrics that MQE knows about, we can say
```
describe all
```

and MQE will spit out all of the metrics that it knows about.

Maybe we're only interested in the ones that have to do with http. We can filter them with a regular expression by adding `match`:

```
describe all match "http"
```

and it will only return the metrics containing a substring matching the regular expression `http`. 

(Everywhere that double-quotes work in MQE, we can also use single-quotes)

Suppose we want to take a look at a metric that comes up, like `http.response_times.ms`. We can ask MQE to describe this metric in particular:

```
describe http.response_times.ms
```

We'll get back a list of columns; each column's name is the name of a tag associated with this metric, and each column will contain all of the tag's possible values. For example, we might get back

| host | datacenter | page   |
|------|------------|--------|
| hs14 | east       | login  |
| hs51 | north      | logout |
| hs78 | west       |        |
| hs98 |            |        |
| hs99 |            |        |

Keep in mind that not every combination of tags listed has to exist. Nor does every tag appear in exactly one series. For example, the actual tag sets might be

* (`hs14`, `east`, `login`)
* (`hs14`, `east`, `logout`)
* (`hs51`, `east`, `login`)
* (`hs51`, `east`, `logout`)
* (`hs78`, `north`, `login`)
* (`hs78`, `north`, `logout`)
* (`hs98`, `west`, `login`)
* (`hs98`, `west`, `logout`)
* (`hs99`, `west`, `login`)
* (`hs99`, `west`, `logout`)

We can restrict the tag sets that we get back by using **predicates**. For example, we might only be interested in series associated with the `west` DC:

```
describe http.response_times.ms
where datacenter = "west"
```

and we'll get back

| host | datacenter | page   |
|------|------------|--------|
| hs98 | west       | login  |
| hs99 |            | logout |

Note that even though we only included `datacenter` in our predicate, the other tag sets changed too! This is because (for example) host `hs14` only has a tag set where `datacenter = "east"`. The actual tag sets associated with this query are only:

* (`hs98`, `west`, `login`)
* (`hs98`, `west`, `logout`)
* (`hs99`, `west`, `login`)
* (`hs99`, `west`, `logout`)

The following also work and do pretty much what you might expect:

```
describe http.response_times.ms
where datacenter != "west"
```

```
describe http.response_times.ms
where host in ("hs78", "hs98", "hs99")
```

```
describe http.response_times.ms
where datacenter = "west" and page = "login"
```

```
describe http.response_times.ms
where datacenter = "west" or host = "hs78"
```

```
describe http.response_times.ms
where datacenter = "west" and not host = "hs99"
```

```
describe http.response_times.ms
where host match "hs9[0-9]*"
```

```
describe http.response_times.ms
where (host match "hs9[0-9]*" or dc = "north") and not host in ("hs99", "hs100")
```

But we probably want to actually see our data, and not just describe it. You can ask MQE to perform a query using `select`.

# Basic Selects

Let's just ask for all of the time series associated with the `http.response_times.ms`:

```
select http.response_times.ms
from -1d to now
```

MQE will present us a graph showing all of the time series associated with `http.response_times.ms`, with one line for each tag set.

We can choose the range of the query by changing the `from` and `to` rules. MQE understands various units for times (`ms`, `s`, `m`, `h` or `hr`, `d`, `w`, `M` or `mo` (28 days)). For example, the following also works:

```
select http.response_times.ms
from -2M to -1w
```

In addition, you can use timestamps in one of various formats, such as 

```
select http.response_times.ms
from 'Fri Jun 21 2016 16:31:15 GMT-0700'
to 'Fri Jun 24 2016 16:31:15 GMT-0700'
```

(if you omit a component of the timerange, in general it will default to 0; if you omit the timezone it will default to UTC)


Just like with `describe`, we can also filter the tag sets we get back using a `where` clause:
```
select http.response_times.ms
where datacenter = 'north'
from -30m to now
```
(note that the `where` clause must go before `from` and `to`)

In this case, we'll only get back lines whose `datacenter` tag is `north`. The available predicates are the same as in `describe`.

# Resolution and `resolution`

Usually, you won't have to specify the resolution of a query: MQE is good at figuring it out for you. If the time interval you're querying needs a certain resolution (because the full resolution data doesn't live along enough), then MQE can change the resolution for you.

In the occasional instances where MQE can't figure it out for you, you can specify the resolution yourself.

```
select http.response_times.ms
where datacenter = 'north'
from -30m to now
resolution 5m
```

The resolution will be at least as coarse as the resolution you specified. If you don't specify one, it implicitly defaults to `30s`.

If MQE ever complains that data doesn't live long enough at the requested resolution, try setting the resolution to be larger than the one you currently have.

# Sampling with `sample by`

When Blueflood performs rollups (taking high-resolution data and sampling it into low-resolution data for long-term storage), it computes a *min*, *mean*, and *max* for each point. By default, MQE with use the `mean` value for each point.

If you want to change this, use `sample by`:

```
select http.response_times.ms
where datacenter = 'north'
from -30m to now
sample by 'max'
```

This is especially useful for metrics like counters (where resets will appear incorrect when sampled by `mean`).

# Selects with Aggregate Functions

MQE provides a large library of functions that you can use to transform, aggregate, and filter your tag sets. Here's a quick walkthrough of some of them.

Let's go back to our HTTP example from before:

```
select http.response_times.ms
from -30m to now
```

MQE might give us back hundreds of time series, one for each page, and host, or any other data associated with each line. But we might just want a quick summary over time. `aggregate.mean` to the rescue!

```
select aggregate.mean(http.response_times.ms)
from -30m to now
```

The *function* `aggregate.mean` will take all of the series returned and [combine them into one series by averaging the lines together](https://github.com/square/metrics/wiki/How-Tags-work-with-Aggregates-and-Joins#aggregation-functions).

We might want to see more information than just the overall mean. For example, how are we doing in each data center? We can answer this question using `group by`:

```
select aggregate.mean(http.response_times.ms group by datacenter)
from -30m to now
```

Now we'll get one line back per datacenter. Note that all other tags will be dropped. We can still use filters to exclude or filter on particular tags. For example

```
select aggregate.mean(http.response_times.ms group by datacenter)
where host != 'really_really_slow_host'
from -30m to now
```

The bad host will be excluded from the aggregate.

If we want to aggregate on multiple features, such as viewing each page's latency by datacenter, we can write

```
select aggregate.mean(http.response_times.ms group by datacenter, page)
where host != 'really_really_slow_host'
from -30m to now
```

Sometimes, there's a tag that we consider unimportant, and we just want to be able to combine lines together when they only differ by it. Let's say that we aren't really interested in division by `host`. Then we can use `collapse by` to write

```
select aggregate.mean(http.response_times.ms collapse by host)
from -30m to now
```

Besides `aggregate.mean`, there's also `aggregate.min`, `aggregate.max`, `aggregate.sum`, `aggregate.count`, and `aggregate.total`. Note that `count` will tell you how many points are present (i.e. not missing) at a point in time, while `total` will tell you how many lines there are (including ones missing points), so `total` will always be constant.

# Selects with Transformation Functions

Transformations can be used to make complex queries in MQE. A common example is wanting to see a rate of change.

For example, suppose we have a metric called `cpu.disk_usage`. We might be interested in how it changes over time. We can query

```
select transform.derivative(cpu.disk_usage)
from -1d to now
```

to see the numerical derivative of `cpu.disk_usage` over time. It will automatically be scaled to units of "X per second", regardless of the resolution the original query was made in (so you don't have to do anything!). For example, if `cpu.disk_usage` is measured in bytes, then the graph will be displayed in `bytes / second` (although MQE doesn't actually label any series with units).

A related transformation is frequently encountered when dealing with **counters**. A counter is a metric that monotonically increases before hitting some maximum value, where it resets. If we used derivative to track these, then we'd see massive negative spikes when the resets occur. To fix this, we'll use the function `transform.rate`, which only considers positive increases to have occurred.

```
select transform.rate(total.response.count)
from -1d to now
```

We can also do manipulations like smoothing data with a moving average:

```
select transform.moving_average(http.response_times.ms, 1h)
from -5d to now
```

MQE will automatically fetch extra data outside your timerange (in this case, `from -5d1h to now`) in order to ensure that the entire moving average is consistent.

# Selects with Joins (`+`, `-`, `*`, `/`)

Sometimes, you have two different series that you want to combine together. For example, you might want to divide the `total.response.time` metric by the `total.response.count` metric to get a "time-per-response" metric. MQE makes this really easy:

```
select total.response.time / total.response.count
from -1d to now
```

Behind the scenes, [MQE perform a tag-based join on the metrics](https://github.com/square/metrics/wiki/How-Tags-work-with-Aggregates-and-Joins#joins). Every line from `total.response.time` will be paired up with every line from `total.response.count`. Then, pairs of lines whose tags don't match will be discarded. The rest are combined together using the operation `/` point-wise.

Because of this, a line in one series may appear several times (or not at all!) in the result of the join. This can be useful, however. Consider this complex query:

```
select aggregate.mean(http.response_times.ms group by host) / aggregate.mean(http.response_times.ms group)
from -1d to now
```

We'll be able to see how each host's response times compare to the overall average at every point in time.

# Select with Filters

Sometimes, you'll have lots of lines in your graph, when you really only care about a few outliers. You can use `filter` functions to reduce the number of lines that MQE returns. For example,

```
select filter.highest_mean(10, http.response_times.ms)
from -1d to now
```

will look at all `http.response_times.ms` lines and sort them by their average value over the past day, then only give you the top 10 highest.

You can also filter by `min` or `max`, and take the `lowest` as well.

Sometimes, after performing one query, you'll want to get a little more context on the series that you've fetched, so you increase the interval that you're querying MQE with. But this can change the series that show up, because their mins, maxes, and means will change as older data is introduced!

To fix this, the `filter` functions take an optional third parameter which specifies how much of the interval to use for the fetch. For example, if I wanted to expand the above query to see these series over the past week, I can write

```
select filter.highest_mean(10, http.response_times.ms, 1d)
from -1w to now
```

# Syntax Sugar

MQE offers some syntax sugar to make it easier to deal with large queries. First, the `select` keyword is optional. So you can write

```
aggregate.mean(http.response_times.ms group by datacenter)
from -1d to now
```

and MQE will know you want to perform a query.

In addition, MQE has a *pipe syntax* for functions. Instead of writing `f(x, y)`, you can always write `x | f(y)`. For example,

```
http.response_times.ms
| aggregate.mean(group by datacenter)
from -1d to now
```

This becomes especially useful when you've got lots of operations in a row:

```
http.responses.count
| transform.rate
| aggregate.sum(group by datacenter)
from -1d to now
```

# Aliases

Sometimes, a complex subexpression can be given a name to make it easier to understand. You can think of these as labels for the data that will be presented (verbatim) in the UI after the series is rendered.

```
(http.responses.count | transform.rate) {HTTP QPS}
+
(rpcs.responses.count | transform.rate) {RPC QPS}
from -1d to now
```

# More Links

* [Function Reference](https://github.com/square/metrics/wiki/Function-Reference)
* [Custom Function Tutorial](https://github.com/square/metrics/wiki/Creating-Custom-Functions)
