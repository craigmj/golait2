package parser

import (
	"fmt"
	"go/ast"
	"reflect"
)

func ExprToJavaType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		for k, v := range map[string]string{
			"byte":    "byte",
			"rune":    "char",
			"uint":    "int",
			"int":     "int",
			"uint8":   "int",
			"uint16":  "int",
			"uint32":  "int",
			"uint64":  "int",
			"int8":    "int",
			"int16":   "int",
			"int32":   "long",
			"int64":   "long",
			"float32": "float",
			"float64": "double",
			"bool":    "boolean",
			"string":  "String",
		} {
			if k == t.Name {
				return v
			}
		}
		return "Object"
	case *ast.StarExpr:
		return ExprToJavaType(t.X)
	case *ast.SelectorExpr:
		return "Object"
	case *ast.MapType:
		return "Object"
	case *ast.ArrayType:
		return ExprToJavaType(t.Elt) + "[]"
	}
	return "Object /* default */"
}

func ExprToJavaJSONType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		for k, v := range map[string]string{
			"byte":    "int",
			"rune":    "int",
			"uint":    "int",
			"int":     "long",
			"uint8":   "int",
			"uint16":  "int",
			"uint32":  "long",
			"uint64":  "long",
			"int8":    "int",
			"int16":   "int",
			"int32":   "long",
			"int64":   "long",
			"float32": "double",
			"float64": "double",
			"bool":    "boolean",
			"string":  "String",
		} {
			if k == t.Name {
				return v
			}
		}
		return "JSONObject"
	case *ast.StarExpr:
		return ExprToJavaType(t.X)
	case *ast.SelectorExpr:
		return "JSONObject /*SelectorExpr */"
	case *ast.MapType:
		return "JSONObject"
	case *ast.ArrayType:
		return "JSONArray"
	}
	return "JSONObject /* default */"
}

// ExprFromJavaJSON returns the method that you would need to call on the
// Java JSONArray to extract the given expr type at index i
func ExprFromJavaJSONArray(expr ast.Expr, i int) string {
	lmap := map[string]string{
		"byte":    "getInt",
		"rune":    "getInt",
		"uint":    "getLong",
		"int":     "getLong",
		"uint8":   "getInt",
		"uint16":  "getInt",
		"uint32":  "getInt",
		"uint64":  "getLong",
		"int8":    "getInt",
		"int16":   "getInt",
		"int32":   "getLong",
		"int64":   "getLong",
		"float32": "getDouble",
		"float64": "getDouble",
		"bool":    "getBoolean",
		"string":  "getString",
	}
	switch t := expr.(type) {
	case *ast.Ident:
		v, ok := lmap[t.Name]
		if !ok {
			v = "getJSONObject"
		}
		return fmt.Sprintf(".%s(%d)", v, i)
	case *ast.ArrayType:
		return fmt.Sprintf(".getJSONArray(%d)", i)
	case *ast.StarExpr:
		return ExprToJavaType(t.X)
	}
	return fmt.Sprintf(".getJSONObject(%d)", i)
}

func ExprToString(expr ast.Expr, pkgName string) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if "" == pkgName {
			return t.Name
		}
		for _, n := range []string{
			"byte",	`error`,
			"rune",
			"uint", "int",
			"uint8", "uint16", "uint32", "uint64",
			"int8", "int16", "int32", "int64",
			"float32", "float64",
			"bool",
			"string",
		} {
			if n == t.Name {
				return n
			}
		}
		return pkgName + "." + t.Name
	case *ast.StarExpr:
		return "*" + ExprToString(t.X, pkgName)
	case *ast.MapType:
		return "map[" + ExprToString(t.Key, pkgName) + "]" + ExprToString(t.Value, pkgName)
	case *ast.SelectorExpr:
		return ExprToString(t.X, "") + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + ExprToString(t.Elt, pkgName)
	case *ast.InterfaceType:
		return "interface{}"
	}
	return reflect.TypeOf(expr).String()
}

func ExprToGoVar(v string, expr ast.Expr, pkgName string) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if "" == pkgName {
			return ``
		}
		for _, n := range []string{
			"byte",	`error`,
			"rune",
			"uint", "int",
			"uint8", "uint16", "uint32", "uint64",
			"int8", "int16", "int32", "int64",
			"float32", "float64",
			"bool",
			"string",
		} {
			if n == t.Name {
				return ``
			}
		}
		return ``
	case *ast.StarExpr:
		return v + `= &` + ExprToString(t.X, pkgName) + `{}`
	case *ast.MapType:
		return v + "= map[" + ExprToString(t.Key, pkgName) + "]" + ExprToString(t.Value, pkgName) +
			"{}"
	case *ast.SelectorExpr:
		// Should never occur
		return `/* SelectorExpr = ` + ExprToString(t.X, "") + "." + t.Sel.Name + `*/`
	case *ast.ArrayType:
		return v + "= []" + ExprToString(t.Elt, pkgName) + "{}"
	case *ast.InterfaceType:
		return "" // v + "= interface{}{}"
	}
	return `/* Should not need to init ` + reflect.TypeOf(expr).String() + ` */`
}

func ExprToPHPType(expr ast.Expr, name string) string {
	switch t := expr.(type) {
	case *ast.Ident:
		for k, v := range map[string]string{
			"byte":    "intval",
			"rune":    "intval",
			"uint":    "intval",
			"int":     "intval",
			"uint8":   "intval",
			"uint16":  "intval",
			"uint32":  "intval",
			"uint64":  "intval",
			"int8":    "intval",
			"int16":   "intval",
			"int32":   "intval",
			"int64":   "intval",
			"float32": "floatval",
			"float64": "floatval",
			"bool":    "boolval",
			"string":  "strval",
		} {
			if k == t.Name {
				return fmt.Sprintf(`%s(%s)`, v, name)
			}
		}
		return name
	case *ast.StarExpr:
		return ExprToPHPType(t.X, name)
	case *ast.SelectorExpr:
		return name
	case *ast.MapType:
		return name
	case *ast.ArrayType:
		return name
	}
	return name
}

func ExprToTypescriptType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		for k, v := range map[string]string{
			"byte":    "string",
			"rune":    "string",
			"uint":    "number",
			"int":     "number",
			"uint8":   "number",
			"uint16":  "number",
			"uint32":  "number",
			"uint64":  "number",
			"int8":    "number",
			"int16":   "number",
			"int32":   "number",
			"int64":   "number",
			"float32": "number",
			"float64": "number",
			"bool":    "boolean",
			"string":  "string",
		} {
			if k == t.Name {
				return v
			}
		}
		return "any"
	case *ast.StarExpr:
		return ExprToTypescriptType(t.X)
	case *ast.SelectorExpr:
		return "any"
	case *ast.MapType:
		return "Map<" + ExprToTypescriptType(t.Key) +
			"," + ExprToTypescriptType(t.Value) + ">"
	case *ast.ArrayType:
		return ExprToTypescriptType(t.Elt) + "[]"
	}
	return "any /* default */"
}
