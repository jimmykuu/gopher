package etc

//go:generate go-bindata -tags "bindata" -ignore "\\.go" -pkg "etc" -o "bindata.go" ../../etc/...
//go:generate go fmt bindata.go
//go:generate rm -f bindata.go.bak
