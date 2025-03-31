#### Setup

Please see home page for a demo setup.

##### Project Structure
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
##### Testing
```
go test ./...
```
##### Committing code
Please ensure the code is correctly formatted and passes the linter.
```
go fmt ./...
```

#### Development Dependencies

MQE uses `github.com/pointlander/peg` as a parser generator. Therefore, it is a dependency to develop MQE (at least to make changes to the MQE query language), but it is not needed to deploy MQE (because the generated parsing files are included in the repository).

