Metrics-indexer
===============

Development
-----------

Check out the project to the development directory.

```
git clone ssh://git@git.corp.squareup.com/vis/metrics-indexer.git $GOPATH/src/square/vis/metrics-indexer/
```

To obtain the list of metrics, you can either:

* query Blueflood's Cassandra (TODO).
* obtain it from a `otsdb2graphite` node.

```
scp alg6.sjc1b:/data/app/otsdb2graphite/metric_list_cache/MetricListFileManager.2015-04-07-15-41-185.txt .
```

* Download Cassandra

We're currently using Cassandra 2.0.X. 2.1.X is unstable and is not
recommended.

Download it from: http://cassandra.apache.org/download/

* To setup schema

```
$CASSANDRA/bin/cqlsh -f schema/schema.cql
```

