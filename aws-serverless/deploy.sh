#!/bin/bash

set -eou pipefail

SANITY="../main.go"
if [ ! -f "${SANITY}" ] ; then
  echo "ERROR: File not found: ${SANITY}" 1>&2
  exit 1
fi

GOOS=linux GOARCH=amd64 go build -o main ..

sam deploy --tags "tenant=fcg" -g


