CURDIR=`pwd`
OLDGOPATH=%$GOPATH%
set GOPATH=%cd%

go install server

set GOPATH=%OLDGOPATH%

echo 'finished'