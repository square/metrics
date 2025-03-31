This reference explains all of the built-in functions for MQE. You can also easily [create your own!](https://github.com/square/metrics/wiki/Creating-Custom-Functions)

## Aggregates

These functions allow `group by` and `collapse by` to specify tags. They'll combine different series in the same expression based on the specified tags.

### `aggregate.max(series [group by tags])`

Max will return the maximum point at each point in time. Missing points will be ignored; if all points are missing at a given point in time, then the result will be missing too.

### `aggregate.mean(series [group by tags])`

Mean returns the mean of all present points at each time. Missing points are ignored; the mean will only include the others. If all points are missing at a given point in time, the result will be missing too.

### `aggregate.min(series [group by tags])`

Min will return the minimum point at each point in time. Missing points will be ignored; if all points are missing at a given point in time, then the result will be missing too.

### `aggregate.sum(series [group by tags])`

Mean returns the mean of all present points at each time. Missing points are ignored; the mean will only include the others. If all points are missing at a given point in time, **the result will be missing, not 0**.

### `aggregate.count(series [group by tags])`

Count tells how many points are present at each point in time.

### `aggregate.total(series [group by tags])`

Total tells how many **series** are present, even if their points are missing. Note that this means it's effectively constant. You can still use it with `group by` and `collapse by` to find the size of various subgroups, however.

## Transforms

Transforms generally operate on a per-line basis. They won't change the number of series returned or their associated tags.

### `transform.derivative(series)`

Derivative gives an instantaneous estimate for the derivative in the form `(y[i+1] - y[i]) / resolution (s)`, so you don't need to worry about your metric's scale or resolution: it will always be in `units/second`.

It's not appropriate for counters, because resets would look like massive negative spikes.

### `transform.rate(series)`

Rate acts just like derivative (including the scaling) except that it suppresses the negative spikes you'd see from counters.

### `transform.integral(series)`

Integral takes in a metric in units of `X/second` and produces a series in units of `X`, measuring the total amount from the start of the time range. It uses a simple Riemann integral over the time series. Missing points are treated as 0 in the input.

### `transform.cumulative(series)`

Cumulative takes in a metric and returns the cumulative sum of the metric. It **does not perform resolution-based scaling**. It is very uncommon that you want to use `cumulative`; generally use `integral` instead. Missing points are treated as 0 in the input.

### `transform.abs(series)`

Abs replaces each point with its absolute value. Missing points remain missing.

### `transform.log(series)`

Log replaces each point with its base-10 logarithm. Missing or non-positive input points will be missing in the result.

### `transform.bound(series, low, high)`

Bound will place limits on the values of `series`. If it goes above the scalar `high` or below the scalar `low`, then those points will be replace by their corresponding bounds. Missing points remain missing. If `high < low` then it will throw an error.

### `transform.lower_bound(series, low)`

Lower bound acts like bound, except with no upper limit.

### `transform.upper_bound(series, high)`

Upper bound acts like bound, except with no lower limit.

### `transform.nan_fill(series, default)`

`NaN` fill replaces all missing points with the specified default scalar.

### `transform.nan_keep_last(series)`

`NaN` keep last will replace each missing point with the most-recent non-missing point. If there is no such point in the current query window, the result will still be missing.

### `transform.moving_average(series, duration)`

Moving average will smooth the series with a moving average of the specified duration (for example, `series | transform.moving_average(1h)`. It automatically widens the underlying query interval in order to be able to smooth the earliest points as well.

Missing points are treated as missing (not 0) in the average; if all of the points in the smoothing interval are missing, the result point will be missing too.

### `transform.exponential_moving_average(series, duration)`

Exponential moving average will smooth the series with an exponential moving average with a halflife of the specified duration (for example, `series | transform.exponential_moving_average(1h)`. The query is widened by one half-life in order to be able to smooth the earliest queried points as well.

Missing points are treated as missing, but the weight of the average still decays over them (so the value will be easier to update as it crosses new points).

### `transform.timeshift(expression, offset)`

Timeshift will evaluate the expression as through it occurred at `offset` time relative to the queried interval. Positive values are future, while negative values are past.

## Tags

### `tag.set(series, tag, value)`

Set will assign the tag in the given series. If it's already present, it will be overwritten. If it's absent, it will be added. You can use this to influence the behavior of aggregates and joins. For example, `tag.set(cpu.percentage, "host", "server12")`.

### `tag.drop(series, tag)`

Drop will remove the tag from the given series. If it's not present, nothing will happen. You can use this to influence the behavior of aggregates and joins. For example, `tag.drop(cpu.percentage, "host")`.

### `tag.copy(series, target, source)`

Copy will copy the value of one tag to another. If the source is not present, the target will be deleted. You can use it to make different metrics more compatible, such as `cpu.user.percentage + tag.copy(box.cpu.percentage, "host", "box")`.

## Forecasts

Simple forecasting functions are available for analyzing your data.

### `forecast.drop(series, duration)`

Drop is helper functions you can use when building or evaluating forecasts. It erases all points in the most recent specified `duration`. 

### `forecast.linear(series [, extra_training_duration])`

Linear will replace each series with its least-squares linear regression. You can specify the `extra_training_duration` to ask it to request more points from earlier in its history.

### `forecast.rolling_seasonal(series, period, learning_rate [, extra_training_duration])`

If your data is approximately seasonal (i.e. cyclic) you can build an estimate for it with `rolling_seasonal`.

Each point in time is estimated using an exponential moving average of the corresponding point in the previous period. The learning rate lies between 0 and 1 and expresses how much weight to place on the current period when updating the rolling model.

You can specify that it should query additional data to build the model by adding the extra duration parameter.

### `forecast.rolling_multiplicative_holt_winters(series, period, level_learning_rate, trend_learning_rate, seasonal_learning_rate [, extra_training_duration])`

The multiplicative Holt-Winters model applies to data that is (approximately) of the form `(a + b*t)*s(t)` where `s` is a seasonal (cyclical) function of known `period`. Each of the parameters can be given its own learning rate (between 0 and 1) which specify the weight of each period over the previous in determining each of the values.

Additionally, you may optionally specify an extra parameter to fetch additional data in order to train the model better.

## Summarize

Summary functions take series and turn them into individual scalars. For example, you might be interested in the average of a series over the query period, or its minimum or maximum.

Summary functions replace *each* line with its summary *scalar*. The resulting scalars are *tagged* using the same tag sets as the original lines.

Tagged scalars can be used wherever a series is expected, and will expand into constants for their series. This is most useful when performing joins (`+`, `-`, `*`, `/`).

However, note that (at this time) joins on two tagged scalars will return a time series, and not a set of tagged scalars.

### `summarize.mean(series [, recent_interval])`

The mean of the series over the queried interval is returned. Missing points are ignored. If all points are missing, the result will be a `NaN` scalar value.

If provided, only the most `recent_interval` of points will be considered.

### `summarize.min(series [, recent_interval])`

The minimum of the series over the queried interval is returned. Missing points are ignored. If all points are missing, the result will be a `NaN` scalar value.

If provided, only the most `recent_interval` of points will be considered.

### `summarize.max(series [, recent_interval])`

The maximum of the series over the queried interval is returned. Missing points are ignored. If all points are missing, the result will be a `NaN` scalar value.

If provided, only the most `recent_interval` of points will be considered.

### `summarize.integral(series)`

The integral of the series over the queried interval is returned. Missing points are treated as 0. The units of the input are interpreted as `X/s` and the result is in units of `X`. Resolution-based scaling occurs automatically.

### `summarize.current(series)`

The current (last or most-recent) point is returned. If it's missing, the result scalar will be `NaN`.

### `summarize.oldest(series)`

The oldest (first or least-recent) point is returned. If it's missing, the result scalar will be `NaN`.

### `summarize.last_not_nan(series)`

The current (last or most-recent) non-missing point is returned. If all points are missing, the result scalar will be `NaN`.

### `summarize.first_not_nan(series)`

The oldest (first or least-recent) non-missing point is returned. If all points are missing, the result scalar will be `NaN`.
