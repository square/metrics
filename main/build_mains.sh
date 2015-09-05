#!/bin/bash -e

current_dir=$(cd "$(dirname "$0")" && pwd)

pushd "$current_dir"
for f in `find * -type d -maxdepth 0`
do
  pushd $f
  go build
  popd
done
popd
