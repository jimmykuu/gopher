// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var fsPublic http.FileSystem = http.Dir("../../static")
	err := vfsgen.Generate(fsPublic, vfsgen.Options{
		PackageName:  "static",
		BuildTags:    "bindata",
		VariableName: "Assets",
		Filename:     "bindata.go",
	})
	if err != nil {
		log.Fatalf("%v", err)
	}
}
