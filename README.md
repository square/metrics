[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/metrics/master/LICENSE)
![Build status](https://travis-ci.org/square/metrics.svg?branch=master)

Metrics
=======

Indexer & Query Engine for Square's metrics.

**This project is still under development and should not be used for anything in production yet. We are not seeking external contributors at this time**

Development
===========

Check out the project to the development directory.

Project Structure
-----------------
```
├── api                # list of publically exposed APIs.
│   └── backend
│       └── blueflood  # implementation of the blueflood backend.
├── assert             # helper functions to make test writing easier.
├── internal           # internal library - should not be exposed to the users.
├── main               # entry point.
│   └── common
├── mocks              # helper code to mock HTTP calls.
├── query              # logic around parsing & execution of the queries.
│   └── aggregate
└── schema             # CQL schema files.
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
golint ./... # TODO - exclude generated files.
```
