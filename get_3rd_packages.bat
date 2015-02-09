set CURDIR=`pwd`
set OLDGOPATH=%$GOPATH%
set GOPATH=%cd%

go get -u github.com/gorilla/mux
go get -u github.com/gorilla/sessions
go get -u gopkg.in/mgo.v2
go get -u code.google.com/p/go-uuid/uuid
go get -u -v code.google.com/p/go.net/websocket
go get -u github.com/jimmykuu/webhelpers
go get -u github.com/jimmykuu/wtforms
go get -u github.com/qiniu/bytes
go get -u github.com/qiniu/rpc
go get -u github.com/qiniu/api
go get -u -v github.com/dchest/captcha
go get -u -v github.com/bradrydzewski/go.auth
go get -u -v github.com/dchest/authcookie
go get -u -v github.com/justinas/nosurf

set GOPATH=%OLDGOPATH%

echo 'finished'
