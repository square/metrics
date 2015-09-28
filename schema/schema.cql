create keyspace metrics_indexer 
with replication = {
  'class':'SimpleStrategy', 'replication_factor': 3
};

use metrics_indexer;

create table metric_names (
  metric_key varchar,
  tag_set varchar,
  primary key ((metric_key), tag_set)
);

create table tag_index (
  tag_key varchar,
  tag_value varchar,
  metric_keys set<varchar>,
  primary key ((tag_key), tag_value)
);

create table metric_name_set (
  shard int,
  metric_names set<varchar>,
  primary key (shard)
);
