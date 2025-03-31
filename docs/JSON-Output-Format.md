MQE queries produce JSON output, available at `/query` by default. The exact format will depend on the query specified in the `query` parameter, but can be checked.

## Describe All

```
{
  "success": true,
  "name": "describe all",
  "body": ["cpu.percent", "cpu.usage", "http.latency", "rpc.latency"], // sorted with natural sort
  "metadata": {
    "count": count, // number of metrics in list
    "profile": profile_data
  }
}
```

## Describe Metrics

```
{
  "success": true,
  "name": "describe metrics",
  "body": ["cpu.percent", "cpu.usage", "http.latency", "rpc.latency"], // sorted with natural sort
  "metadata": {
    "count": count, // number of metrics in list
    "profile": profile_data
  }
}
```

## Describe

```
{
  "success": true,
  "name": "describe",
  "body": {
    "dc": ["east", "north", "south", "west"],
    "host": ["server7", "server8", "server9", "server10"] // sorted with natural sort
  },
  "metadata": {
    "profile": profile_data
  }
}
```

## Select

```
{
  "success": true,
  "name": "select",
  "body": [
    {
      "query": "cpu.percentage[app = 'mqe'] {CPU}",
      "name": "CPU",
      "type": "series",
      "series": [
        {"tagset": {"app": "mqe", "dc": "east"}, "values": [20, 21, 22, 19, 18]},
        {"tagset": {"app": "mqe", "dc": "west"}, "values": [19, 19, 18, 19, null]},
        {"tagset": {"app": "mqe", "dc": "north"}, "values": [null, null, null, null, null]},
        {"tagset": {"app": "mqe", "dc": "south"}, "values": [23, 24, 20, 19, 18, null]},
      ],
      "timerange": {
        "start":1468966410000,
        "end":1468966530000,
        "resolution":30000
      }
    }
  ],
  "metadata": {
    "description": { "app": ["mqe"], "dc": ["east", "west", "north", "south"] },
    "notes": null, // or, ["foo", "bar"]
    "profile": profile_data
  }
}
```

## Error

```
{
  "success": false,
  "message": "No such metric with name `fodfod`"
}
```

```
{
  "success": false,
  "message": "performing fetch of 3433 additional series brings the total to 6866, which exceeds the specified limit 5000"
}
```


TODO: `profile_data` specification and elaboration

TODO: scalar values