#!/usr/bin/env bash

(
  cd "$( dirname "${BASH_SOURCE[0]}" )"
  peg -inline -switch language.peg
  goimports -w ./language.peg.go # format the file; optional
)
