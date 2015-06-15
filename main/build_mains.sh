#!/bin/bash -ex

current_dir=$(cd "$(dirname "$0")" && pwd)

pushd "$current_dir"
for f in *.go
do
  go build "$current_dir/$f"
done
popd
