#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go install server
go install movetocontents

export GOPATH="$OLDGOPATH"

echo 'finished'
