package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestImport(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(),
		"test",
		`package test
import (
	a "one.two.three"
	b "two"
)
type Test struct {}
func (*Test) A(int, a.One) b.Two {
	return &b.Two{}
}
func (*Test) B(a.Two, b.Three) error {
	return nil
}
`, 0)
	if nil != err {
		t.Error(err)
	}
	imports := NewImportList(file)
	for _, decl := range file.Decls {
		switch e := decl.(type) {
		case *ast.FuncDecl:
			imports.AddFunction(e)
		}
	}
	for _, i := range imports.Imports {
		t.Log(i)
	}
	im := imports.getImport("", `"one.two.three"`)
	if USE_PARAM != im.Usage {
		t.Errorf("Import of one.two.three expected PARAM usage only, but got %d: %q", im.Usage, *im)
	}
	im = imports.getImport("", `"two"`)
	if (USE_RESULT | USE_PARAM) != im.Usage {
		t.Errorf("Import of 'two' expected usage RESULT|PARAM, but got %d: %q", im.Usage, *im)
	}
}
