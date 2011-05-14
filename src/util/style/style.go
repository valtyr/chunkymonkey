package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type PackageVisitor struct {}

func (v *PackageVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.TypeSpec:
		w = &TypeVisitor{
			typeSpec: n,
		}
	default:
		w = v
	}
	return
}

type TypeVisitor struct {
	typeSpec *ast.TypeSpec
}

func (v *TypeVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.ArrayType:
	case *ast.ChanType:
	case *ast.FuncType:
	case *ast.InterfaceType:
		fmt.Printf("node inside type spec %q: %#v\n", v.typeSpec.Name.Name, n)
	case *ast.MapType:
	case *ast.StructType:
	default:
		return v
	}
	return
}

func errPrintf(format string, args... interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	flag.Parse()
	filenames := flag.Args()
	fset := token.NewFileSet()
	pkgMap, firstErr := parser.ParseFiles(fset, filenames, 0)
	if firstErr != nil {
		errPrintf("Error while parsing: %v\n", firstErr)
	}
	v := new(PackageVisitor)
	for pkgName, pkg := range pkgMap {
		ast.Walk(v, pkg)
		fmt.Printf("package %q found\n", pkgName)
	}
}
