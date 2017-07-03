package static

//go:generate go-bindata -tags "bindata" -ignore "\\.go|\\.less" -pkg "static" -o "bindata.go" ../../static/...
//go:generate go fmt bindata.go
//go:generate rm -f bindata.go.bak
