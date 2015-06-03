#!/usr/bin/env bash

# a hack to determine the location of this script:
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

peg -inline -switch $DIR/language.peg
