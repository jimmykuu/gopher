set CURDIR=`pwd`
set OLDGOPATH=%$GOPATH%
set GOPATH=%cd%

gofmt -w src/gopher

go install server

set GOPATH=%OLDGOPATH%

echo 'finished'