package templates

//go:generate go-bindata -tags "bindata" -ignore "\\.go" -pkg "templates" -o "bindata.go" ../../templates/...
//go:generate go fmt bindata.go
//go:generate rm -f bindata.go.bak
