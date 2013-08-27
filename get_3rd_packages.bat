set CURDIR=`pwd`
set OLDGOPATH=%$GOPATH%
set GOPATH=%cd%

go get -u github.com/gorilla/mux
go get -u github.com/gorilla/sessions
go get -u labix.org/v2/mgo
go get -u code.google.com/p/go-uuid/uuid
go get -u github.com/jimmykuu/webhelpers
go get -u github.com/qiniu/bytes
go get -u github.com/qiniu/rpc
go get -u github.com/qiniu/api

set GOPATH=%OLDGOPATH%

echo 'finished'