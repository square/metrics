[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/metrics/master/LICENSE)
[![Build Status](https://travis-ci.org/square/metrics.svg?branch=master)](https://travis-ci.org/square/metrics)

#### Metrics Query Engine

Metrics Query Engine(MQE) provides SQL-like interface to time series data with powerful functions.

For example, to find which 10 endpoints have the highest HTTP latency on your web application farm:

```
select connection.http.latency
| aggregate.sum(group by endpoint)
| filter.highest_mean(10)
where application = 'httpd'
from -2hr to now
```

##### Why
Square generates approximately 2.5 million metrics (as of July 2015). The large volume of unstructured metric names makes it difficult to search for and discover metrics relevant to a particular host, app, service, connection type, or data center. Metrics Query Engine uses tagged metrics as a way to structure metric names so that they can be more easily queried and discovered.


###### See wiki for installation, setup and development. 
