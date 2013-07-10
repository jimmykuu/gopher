#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go get -u github.com/gorilla/mux
go get -u github.com/gorilla/sessions
go get -u labix.org/v2/mgo
go get -u code.google.com/p/go-uuid/uuid
go get -u github.com/jimmykuu/webhelpers
go get -u github.com/jimmykuu/wtforms
go get -u github.com/qiniu/bytes
go get -u github.com/qiniu/rpc
go get -u github.com/qiniu/api

export GOPATH="$OLDGOPATH"

echo 'finished'