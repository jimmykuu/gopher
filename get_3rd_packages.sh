#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go get -u -v github.com/gorilla/mux
go get -u -v github.com/gorilla/sessions
go get -u -v labix.org/v2/mgo
go get -u -v code.google.com/p/go-uuid/uuid
go get -u -v github.com/jimmykuu/webhelpers
go get -u -v github.com/jimmykuu/wtforms
go get -u -v github.com/qiniu/bytes
go get -u -v github.com/qiniu/rpc
go get -u -v github.com/qiniu/api
go get -u -v github.com/dchest/captcha

export GOPATH="$OLDGOPATH"

echo 'finished'
