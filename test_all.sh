#!/bin/bash -e

current_dir=$(cd "$(dirname "$0")" && pwd)

pushd "$current_dir"
for f in `find * -type d`
do
  pushd $f > /dev/null
  echo "Testing $f"
  ls *go > /dev/null 2>&1 && go test
  popd > /dev/null
done
popd
