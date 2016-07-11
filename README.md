[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/metrics/master/LICENSE)
[![Build Status](https://travis-ci.org/square/metrics.svg?branch=master)](https://travis-ci.org/square/metrics)

#### Metrics Query Engine [(Version 1.0)](https://github.com/square/metrics/releases/tag/v1.0)

```
go get "github.com/square/metrics"
```

Metrics Query Engine(MQE) provides SQL-like interface to time series data with powerful functions to aggregate, filter and analyze.

For example, to find which 10 endpoints have the highest HTTP latency on your web application farm:

```
select connection.http.latency
| aggregate.sum(group by endpoint)
| filter.highest_mean(10)
where application = 'httpd'
from -2hr to now
```

Or maybe you want to compare cpus used vs allocated across your cluster for a particular application

```
inspect.cgroup.cpustat.usage | aggregate.sum,
inspect.cgroup.cpustat.total | aggregate.sum
where service match 'blueflood'
from -10m to now
```

Or you want to see how many cumulative seconds have been spent serving an API request.

```
transform.integral(
 aggregate.sum(transform.rate(`framework.actions.service-api.response_codes.X00`[type='200'])
 *
`framework.actions.service-api.response_times.histogram`[distribution='mean'])
)

where app = 'secretapp' and service = 'SecretService' and api = 'GetSecret'

from -1w to now
```

##### Why
Square collects millions of signals every few seconds from application servers and datacenters. The large volume of unstructured metric names makes it difficult to search for and discover metrics relevant to a particular host, app, service, connection type, or data center. Metrics Query Engine uses tagged metrics as a way to structure metric names so that they can be more easily queried and discovered.

#### Go Version

MQE supports Go 1.5 and up.


###### See wiki for installation, setup and development.
