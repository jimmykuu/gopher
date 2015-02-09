#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

go get -u -v github.com/gorilla/mux
go get -u -v github.com/gorilla/sessions
go get -u -v gopkg.in/mgo.v2
go get -u -v code.google.com/p/go-uuid/uuid
go get -u -v code.google.com/p/go.net/websocket
go get -u -v github.com/jimmykuu/webhelpers
go get -u -v github.com/jimmykuu/wtforms
go get -u -v github.com/qiniu/bytes
go get -u -v github.com/qiniu/rpc
go get -u -v github.com/qiniu/api
go get -u -v github.com/dchest/captcha
go get -u -v github.com/bradrydzewski/go.auth
go get -u -v github.com/dchest/authcookie
go get -u -v github.com/justinas/nosurf

export GOPATH="$OLDGOPATH"

echo 'finished'
