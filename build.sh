#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go install server

export GOPATH="$OLDGOPATH"

echo 'finished'
