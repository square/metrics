[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/metrics/master/LICENSE)
[![Build Status](https://travis-ci.org/square/metrics.svg?branch=master)](https://travis-ci.org/square/metrics)

Metrics
=======

Indexer & Query Engine for Square's metrics.

**This project is still under development and should not be used for anything in production yet. We are not seeking external contributors at this time**

We currently support Go 1.4 and Go 1.5.

Development
===========

Check out the project to the development directory.

Project Structure
-----------------
```
Main packages:
├── api                 # core type and function definitions
├── function            # MQE function definition interface
│   └── registry        # registry for custom MQE functions
├── main
│   └── ui              # the UI executable
├── metric_metadata     # interface for storing metric metadata
│   └── cassandra       # Cassandra backend for metric metadata
├── query               # query language and parsing
├── timeseries_storage  # interface for storing time series data

Miscellaneous packages:
├── compress            # experimental metrics-compression protocol
├── inspect             # profiling to measure MQE query performance
├── log                 # custom logging
├── optimize            # MQE optimization interface
├── schema              # example Cassandra schema configurations
├── testing_support     # mocks for interfaces
├── ui                  # webserver for UI interface
├── util                # conversion rules for graphite metrics
```

Cassandra
---------

We're currently using Cassandra 2.0.X. 2.1.X is unstable and is not
recommended.

Download it from: http://cassandra.apache.org/download/

* To setup schema

```
# Production schema
$CASSANDRA/bin/cqlsh -f schema/schema.cql
# Testing Schema
$CASSANDRA/bin/cqlsh -f schema/schema_test.cql
```

Dependencies
------------

```
go get github.com/gocql/gocql
go get github.com/pointlander/peg
go get gopkg.in/yaml.v2
```

Testing
-------

```
go test ./...
```

Committing code
---------------

Please ensure the code is correctly formatted and passes the linter.

```
go fmt ./...
```
