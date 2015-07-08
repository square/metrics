# Examples

### Describe All

List all metric names in the system:

```
describe all
```

### Describe

Find what tagsets are associated to a given metric name:

```
describe cpu
```

If the name contains special characters (any character other than a letter, number, or period) or does not begin with a letter, then enclose the metric name in backticks:

```
describe `cpu-userspace`
```

### Describe Metrics

To find every metric that has a given tag (key, value) pair, use the `describe metrics` command:

```
describe metrics where app = 'metrics-indexer'
```

Only one (key, value) pair can be selected on at a time.

### Select

Find how much CPU each app is using in the past 4 hours:

```
select cpu from -4hr to now
```

The `select` keyword is optional:

```
cpu from -4hr to now
```

We might want to see the total amount of CPU usage everywhere:

```
aggregate.sum( cpu ) from -4hr to now
```

We can use pipes for brevity:

```
cpu | aggregate.sum from -4hr to now
```

We might want to see how much each app is using:

```
aggregate.sum( cpu group by app ) from -4hr to now
```

Or with pipes:

```
cpu | aggregate.sum(group by app) from -4hr to now
```

We can see how much each app takes in each datacenter:

```
cpu | aggregate.sum(group by app, datacenter) from -4hr to now
```

And we can see what percent of each datacenter each app consumes:

```
cpu | aggregate.sum(group by app, datacenter) / (cpu | aggregate.sum(group by datacenter)) * 100 from -4hr to now
```

Or in function notation:

```
aggregate.sum(cpu group by app, datacenter) / aggregate.sum(cpu group by datacenter) * 100
```

We might want to find how much cpu is used by userspace and say by kernel on every host:

```
cpu.kernel + cpu.user from -4hr to now
```

Although we might want to only restrict this to a small number of apps:

```
cpu.kernel + cpu.user
where app in ('metrics-query-engine', 'blueflood', 'cassandra')
from -4hr to now
```

To look for only the hosts consuming the most cpu for these apps, use `filter_highest.max`:

```
cpu.kernel + cpu.user
| filter.highest_max(10)
where app in ('metrics-query-engine', 'blueflood', 'cassandra')
from -4hr to now
```
