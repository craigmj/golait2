package parser

import (
	"errors"
	"go/ast"
	"path/filepath"
	"strings"

	// "github.com/golang/glog"
)

const (
	NONE = 1 << iota
	USE_PARAM
	USE_RESULT
)

/* ImportList contains the list of all packages that are imported
 * by the root package and that are required in the various
 * methods that we examine.
 */
type Import struct {
	Name  string // Name of the package in the root package
	Path  string // Path to the package in the root package
	Usage int    // USE_PARAM, USE_RESULT, NONE
}

func (i *Import) IsParam() bool {
	return 0 != (i.Usage & USE_PARAM)
}

func (i *Import) IsResult() bool {
	return 0 != (i.Usage & USE_RESULT)
}

func (i *Import) ShouldImport() bool {
	return 0 == i.Usage || (i.Usage&USE_PARAM == USE_PARAM)
}

func (i *Import) ShouldImportWithReturns() bool {
	return 0==i.Usage || (i.Usage&USE_PARAM==USE_PARAM) ||
		(i.Usage&USE_RESULT==USE_RESULT)
}

type ImportList struct {
	imports []*ast.ImportSpec
	Imports []*Import
}

func NewImportList(file *ast.File) *ImportList {
	return &ImportList{
		imports: file.Imports,
		Imports: make([]*Import, 0),
	}
}

func (i *ImportList) AddFunction(e *ast.FuncDecl) error {
	err := i.addFieldList(e.Type.Params, USE_PARAM)
	if nil != err {
		return err
	}
	return i.addFieldList(e.Type.Results, USE_RESULT)
}

func (i *ImportList) AddImports(imports ...string) string {
	for _, imp := range imports {
		i.getImport("", `"`+imp+`"`)
	}
	return ""
}

// Add will add an ast.Expr into the ImportsList.
// It adds any parts of the expression, whether it is a
// *, [] or map, or just a simple expression
func (i *ImportList) Add(e ast.Expr, use int) error {
	switch t := e.(type) {
	case *ast.ArrayType:
		return i.Add(t.Elt, use)
	case *ast.MapType:
		err := i.Add(t.Key, use)
		if nil != err {
			return err
		}
		return i.Add(t.Value, use)
	case *ast.Ident:
		return nil
	case *ast.SelectorExpr:
		// t.Sel *Ident is the field selector
		// Find the pkg in the file's imports
		sel := ExprToString(t.X, "")
		for _, imp := range i.imports {
			path := imp.Path.Value
			// Trim leading and trailing Inverted commas
			path = strings.Trim(path, ` "'`+"`")
			alias := filepath.Base(path)
			if nil != imp.Name {
				alias = imp.Name.Name
			}
			if alias == sel {
				// We trim the trailing
				im := i.getImport(alias, imp.Path.Value)
				im.Usage |= use
				return nil
			}
		}
		return errors.New("Failed to resolve selected package " +
			t.Sel.Name + ".")
	case *ast.StarExpr:
		return i.Add(t.X, use)
	case *ast.InterfaceType:
		return nil
	}
	return nil
}

func (i *ImportList) getImport(name, path string) *Import {
	for _, im := range i.Imports {
		if im.Path == path {
			return im
		}
	}
	im := &Import{Name: name, Path: path}
	i.Imports = append(i.Imports, im)
	return im
}

func (i *ImportList) addFieldList(e *ast.FieldList, use int) error {
	var err error
	for _, f := range e.List {
		if err = i.Add(f.Type, use); nil != err {
			return err
		}
	}
	return nil
}
