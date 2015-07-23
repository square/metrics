-- Create keyspace
create keyspace metrics_indexer_test with replication = {'class':'SimpleStrategy', 'replication_factor': 1};

use metrics_indexer_test;

-- metric_names
create table metric_names (
  metric_key varchar,
  tag_set varchar,
  primary key ((metric_key), tag_set)
);

-- tag_index
create table tag_index (
  tag_key varchar,
  tag_value varchar,
  metric_keys set<varchar>,
  primary key ((tag_key), tag_value)
);

-- metric_name_set
create table metric_name_set (
  shard int,
  metric_names set<varchar>,
  primary key (shard)
);

-- graphite_names
create table graphite_names (
  shard int,
  graphite_name varchar,
  primary key ((shard), graphite_name)
);
