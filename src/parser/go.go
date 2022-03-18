package parser

import (
	"fmt"
	"go/ast"
	"strings"
)

// "text/template"

type ClassDefinition struct {
	PackageName     string // The name of the package for our RPC classes
	RootPackage     string // The name of the package where the API object is
	RootPackagePath string // Path to the root package
	Constructor     string // The name of the constructor function for an API object, if there is one
	ClassName       string // The name of the Type for the API object
	Recover         bool   // Should we auto-recover from panics

	ConnectionClass            string // The name of the class to use for Connections
	ConnectionClassConstructor string // The New method to create a ConnectionClass
	// The ClassName is the Context

	Imports *ImportList         // Imports required in our function signatures
	Methods []*MethodDefinition // All the methods in the Class
	JsApply bool                // JsApply uses the .apply calling method for js calls
	JsExport bool 				// JsExport if we should add 'export' to our class definitions
}

// NewClassDefinition returns a new class defined from the
// methods in the given file
func NewClassDefinition(file *ast.File, rootPackagePath, packageName, className, constructor string, recoverFlag bool, connectionClass, connectionConstructor string) (*ClassDefinition, error) {
	class := &ClassDefinition{
		PackageName:                packageName,
		RootPackage:                file.Name.Name,
		RootPackagePath:            rootPackagePath,
		Constructor:                constructor,
		ClassName:                  className,
		Recover:                    recoverFlag,
		Imports:                    NewImportList(file),
		Methods:                    make([]*MethodDefinition, 0),
		ConnectionClass:            connectionClass,
		ConnectionClassConstructor: connectionConstructor,
	}
	for _, decl := range file.Decls {
		// We skip all functions and any methods that aren't relevant
		// to our class
		f, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// Check that this is a method of our API class
		if nil == f.Recv || "*"+className != ExprToString(f.Recv.List[0].Type, "") {
			continue
		}
		if !ast.IsExported(f.Name.Name) {
			continue
		}
		class.Methods = append(class.Methods, NewMethodDefinition(f))
		err := class.Imports.AddFunction(f)
		if nil != err {
			return nil, err
		}
	}
	return class, nil
}

type MethodDefinition struct {
	Abbreviation string // Abbreviation of the method if used
	Name         string // Name of the method
	Parameters   []*ParameterDefinition
	Results      ResultsDefinition
}

func NewMethodDefinition(f *ast.FuncDecl) *MethodDefinition {
	m := &MethodDefinition{
		Name:       f.Name.Name,
		Parameters: make([]*ParameterDefinition, 0),
	}
	// A parameter can have multiple names, as in func(a,b,c int)
	// So we need to convert this into multiple Parameter types
	for _, p := range f.Type.Params.List {
		for nameIndex := range p.Names {
			m.Parameters = append(m.Parameters, &ParameterDefinition{NewAstField(p, nameIndex)})
		}
	}
	m.Results.fields = f.Type.Results
	return m
}

func (m *MethodDefinition) ResultsJavaFields() string {
	java := make([]string, m.Results.fields.NumFields())
	for i, a := range m.Results.fields.List {
		java[i] = fmt.Sprintf("p%d %s", i+1, ExprToJavaJSONType(a.Type))
	}
	return strings.Join(java, ", ")
}

func (m *MethodDefinition) ResultsJavaValues(varr string) string {
	java := make([]string, m.Results.fields.NumFields())
	for i, a := range m.Results.fields.List {
		java[i] = fmt.Sprintf("%s%s", varr, ExprFromJavaJSONArray(a.Type, i))
	}
	return strings.Join(java, ", ")
}

func (m *MethodDefinition) ParameterNameList() []string {
	a := make([]string, len(m.Parameters))
	for i, p := range m.Parameters {
		a[i] = p.Field.Name()
	}
	return a
}

func (m *MethodDefinition) ParameterTypedNameListTs() []string {
	a := make([]string, len(m.Parameters))
	for i, p := range m.Parameters {
		a[i] = fmt.Sprintf("%s:%s", p.Field.Name(), ExprToTypescriptType(p.Field.Type))
	}
	return a
}

func (m *MethodDefinition) ParametersLen() int {
	return len(m.Parameters)
}

type ParameterDefinition struct {
	Field *AstField
}

type AstField ast.Field

func NewAstField(f *ast.Field, nameIndex int) *AstField {
	nf := &ast.Field{
		Doc:     f.Doc,
		Names:   []*ast.Ident{f.Names[nameIndex]},
		Type:    f.Type,
		Tag:     f.Tag,
		Comment: f.Comment,
	}
	return (*AstField)(nf)
}

func (a *AstField) Name() string {
	if nil == a.Names || 0 == len(a.Names) {
		return ""
	}
	return a.Names[0].Name
}

func (a *AstField) TitleName() string {
	return strings.ToTitle(a.Name())
}

func (a *AstField) GoType(pkgName string) string {
	return ExprToString(a.Type, pkgName)
}

func (a *AstField) JavaType() string {
	return ExprToJavaType(a.Type)
}

func (a *AstField) CoerceInPHP() string {
	return ExprToPHPType(a.Type, `$`+a.Name())
}

type ResultsDefinition struct {
	fields *ast.FieldList
}

func (r *ResultsDefinition) Length() int {
	return len(r.fields.List)
}

func (r *ResultsDefinition) Fields() []*ast.Field {
	return r.fields.List
}

func (r *ResultsDefinition) LengthArray() []int {
	arr := make([]int, r.Length())
	for i := 0; i < r.Length(); i++ {
		arr[i] = i
	}
	return arr
}

func (r *ResultsDefinition) LastElementIndex() int {
	return r.Length() - 1
}

func (r *ResultsDefinition) IsErrorLast() bool {
	if 0 == r.Length() {
		return false
	}
	return "error" == ExprToString(r.fields.List[r.LastElementIndex()].Type, "")
}
