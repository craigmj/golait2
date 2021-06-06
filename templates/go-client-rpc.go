package {{.PackageName}}

import (
{{with .Imports -}}
	{{range .Imports }}{{if .ShouldImportWithReturns -}}
	{{.Name}} {{.Path}}{{end}}
	{{end -}}
{{end}}

	`github.com/ethereum/go-ethereum/rpc`
	`github.com/juju/errors`
)

type {{.ConnectionClass}} struct {
	server *rpc.Client
}

func New{{.ConnectionClass}}(server string) (*{{.ConnectionClass}}, error) {
	s, err := rpc.DialHTTP(server)
	if nil!=err {
		return nil, err
	}
	return &{{.ConnectionClass}}{server:s}, nil
}

func (c *{{.ConnectionClass}}) Close() {
	c.server.Close()
}

func With{{.ConnectionClass}}(server string, f func (*{{.ConnectionClass}}) error) error {
	c, err := New{{.ConnectionClass}}(server)
	if nil!=err {
		return errors.Trace(err)
	}
	defer c.Close()
	return f(c)
}

func (c *{{$.ConnectionClass}}) call(out interface{}, method string, args ...interface{}) error {
	ret := []interface{}{out}
	if err := c.server.Call(&ret, method, args...); nil!=err {
		return errors.Trace(err)
	}
	return nil
}

{{range .Methods}}
func (_c *{{$.ConnectionClass}}) {{.Name}}(
	{{range $i,$p := .Parameters -}}
			{{$p.Field.Name}} {{$p.Field.GoType "_root"}},
	{{end}}) (
		{{- range $i, $r := .Results.Fields -}}
			r{{$i}} {{ExprToString $r.Type "_root"}},
		{{- end}}) {
	{{ if eq 0 .Results.Length }}
	err := _c.call(nil, `{{.Name}}`, {{.ParameterNameList | Join ", " }})
	if nil!=err {
		log.Errorf(err)
	}
	{{ else if eq 1 .Results.Length}}
	r0 = _c.call(nil, `{{.Name}}`, {{.ParameterNameList | Join ", " }})
	{{ else if eq 2 .Results.Length}}{{with index .Results.Fields 0}}
	{{ExprToGoVar `r0` .Type "_root"}}
	{{end}}
	r{{.Results.LastElementIndex}} = _c.call(&r0, `{{.Name}}`, {{.ParameterNameList | Join ", " }})
	{{ else }}
	panic("{{$.ClassName}} method {{.Name}} returns {{.Results.Length}} values, but maximum of 2 supported");
	{{end}}
	return
}
{{end}}
