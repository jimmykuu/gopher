#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go install server
go install gravatar2qiniu

export GOPATH="$OLDGOPATH"

echo 'finished'
