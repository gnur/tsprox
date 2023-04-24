#!/bin/bash

function log {
  echo "[$(date)]: $*"
}

export HOSTNAME="tsprox-dev"
export VERBOSE="true"

go run main.go
