[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/metrics/master/LICENSE)
[![Build Status](https://travis-ci.org/square/metrics.svg?branch=master)](https://travis-ci.org/square/metrics)

#### Metrics Query Engine

Metrics Query Engine(MQE) provides SQL-like interface to time series data with powerful functions to aggregate, filter and analyze.

For example, to find which 10 endpoints have the highest HTTP latency on your web application farm:

```
select connection.http.latency
| aggregate.sum(group by endpoint)
| filter.highest_mean(10)
where application = 'httpd'
from -2hr to now
```

Or maybe you want to find application CPU usage vs allocated across your cluster.

```
inspect.cgroup.cpustat.usage | aggregate.sum,
inspect.cgroup.cpustat.total | aggregate.sum
where service match 'blueflood'
from -10m to now
```


##### Why
Square collects millions of signals from application servers and datacenters. The large volume of unstructured metric names makes it difficult to search for and discover metrics relevant to a particular host, app, service, connection type, or data center. Metrics Query Engine uses tagged metrics as a way to structure metric names so that they can be more easily queried and discovered.


###### See wiki for installation, setup and development. 
