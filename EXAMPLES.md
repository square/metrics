# Examples

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

We might want to find how much memory our users and our kernels are using on every host:

```
memory.kernel + memory.user from -4hr to now
```

Although we might want to only restrict this to a small number of apps:

```
memory.kernel + memory.user
where app in ('metrics-query-engine', 'blueflood', 'cassandra')
from -4hr to now
```

To look for only the hosts consuming the most memory, use `filter_highest.max`:

```
memory.kernel + memory.user
| filter_highest.max(10)
where app in ('metrics-query-engine', 'blueflood', 'cassandra')
from -4hr to now
```
