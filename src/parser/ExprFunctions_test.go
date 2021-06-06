package parser

import (
	"go/parser"
	"testing"
)

func TestExprToString(t *testing.T) {
	for _, test := range []string{
		"one.two.three",
		"int",
		"*test.One",
		"map[string][]*string",
		"x",
	} {
		expr, err := parser.ParseExpr(test)
		if nil != err {
			t.Errorf("Error parsing expr %s: %s", expr, err.Error())
		}
		res := ExprToString(expr, "")
		if test != res {
			t.Errorf("Expected " + test + ", got " + res)
		}
	}

	for _, test := range [][2]string{
		[2]string{"int", "int"},
		[2]string{"*test.One", "*test.One"},
		[2]string{"Test", "_root.Test"},
	} {
		expr, err := parser.ParseExpr(test[0])
		if nil != err {
			t.Errorf("Error parsing expr %s : %s", expr, err.Error())
			continue
		}
		res := ExprToString(expr, "_root")
		if test[1] != res {
			t.Errorf("Expected " + test[1] + ", got " + res)
		}
	}
}
