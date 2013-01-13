#!/usr/bin/env bash

if [ ! -f make ]; then
echo 'make must be run within its container folder' 1>&2
exit 1
fi

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go install server

export GOPATH="$OLDGOPATH"

echo 'finished'