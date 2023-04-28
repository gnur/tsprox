#!/bin/bash

function log {
  echo "[$(date)]: $*"
}

export HOSTNAME="tsprox-dev"
export VERBOSE="true"
export PROXY_HOST="http://localhost:8008"
export ENABLE_CAPTURE=false
export MAX_CAPTURES=10

go run .
