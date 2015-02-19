package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/gopherjs/gopherjs/compiler"
)

func main() {
	packages := make(map[string]*compiler.Archive)
	var pkgsToLoad []string
	importContext := compiler.NewImportContext(func(path string) (*compiler.Archive, error) {
		if pkg, found := packages[path]; found {
			return pkg, nil
		}
		pkgsToLoad = append(pkgsToLoad, path)
		return &compiler.Archive{}, nil
	})
	fileSet := token.NewFileSet()
	pkgsReceived := 0
	file, err := parser.ParseFile(fileSet, "prog.go", []byte("package main"), parser.ParseComments)
	if err != nil {
		fmt.Println(err)
	}
	mainPkg, err := compiler.Compile("main", []*ast.File{file}, fileSet, importContext, false)
	packages["main"] = mainPkg
	if err != nil && len(pkgsToLoad) == 0 {
		fmt.Println(err)
	}
	var allPkgs []*compiler.Archive
	if len(pkgsToLoad) == 0 {
		allPkgs, _ = compiler.ImportDependencies(mainPkg, importContext.Import)
	}
	if len(pkgsToLoad) != 0 {
		pkgsReceived = 0
		for _, p := range pkgsToLoad {
			path := p

		}
	}
}
