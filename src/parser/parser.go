package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/golang/glog"
)

func processFuncDecl(f *ast.FuncDecl) error {
	recv := "nil"
	if nil != f.Recv {
		recv = ExprToString(f.Recv.List[0].Type, "")
	}
	fmt.Println("("+recv+")", f.Name)
	return nil
}

func Parse(in string) error {
	fileset := token.NewFileSet()

	source, err := parser.ParseFile(fileset, in, nil, 0 /* Parse everything */)
	if nil != err {
		glog.Fatal(err)
	}

	glog.Infof("Parsed file %s fine", in)
	glog.Infof("File %s is for package %s", in, source.Name.String())

	for _, decl := range source.Decls {
		switch f := decl.(type) {
		case *ast.FuncDecl:
			processFuncDecl(f)
		}
	}
	return nil
}
