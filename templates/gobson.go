package {{.PackageName}}
{{$packageName := .PackageName}}

{{/*
The 'execute' template should be called on a MethodDefinition
object, and will write out the code to execute the method itself.
*/}}
{{define "execute"}}
context.{{.Name}}(
		{{range $i, $p := .Parameters}}
			args.{{$p.Field.TitleName}},
		{{end }}
				)
{{end}}

import (
{{with .Imports}}
	{{.AddImports "errors" "fmt" "io" "net/http" "github.com/golang/glog" "bytes" "gopkg.in/mgo.v2/bson"}}
	{{range .Imports}}{{if .ShouldImport}}
	{{.Name}} {{.Path}}{{end}}{{end}}
{{end}}

	_root "{{.RootPackagePath}}"
)


/**
 * HttpHandlerFunc is the entry point for the http BSON POST
 * handler.
 */
func HttpBsonHandlerFunc(w http.ResponseWriter, r *http.Request) {
	{{if .Constructor}}
	context := _root.{{.Constructor}}(w,r)
	{{else}}
	context := &_root.{{.ClassName}}{}
	{{end}}
	processBsonRpc(r.Body, w, context)
}

const BSONRPC_ERROR_PARSE_ERROR = -32700
const BSONRPC_ERROR_INVALID_REQUEST = -32600
const BSONRPC_ERROR_METHOD_NOT_FOUND = -32601
const BSONRPC_ERROR_INVALID_PARAMS = -32602
const BSONRPC_ERROR_INTERNAL_ERROR = -32603
const BSONRPC_ERROR_APPLICATION_ERROR = -1000


type bsonRequest struct {
	Jsonrpc string `bson:"jsonrpc"`
	Method string `bson:"method"`
	Params bson.Raw `bson:"params"`
	Id	interface{} `bson:"id"`
}
type bsonError struct {
	Code int `bson:"code"`
	Message string `bson:"message"`
	Data interface{} `bson:"data,omitempty"`
}
type bsonResponse struct {
	Jsonrpc string `bson:"jsonrpc"`
	Id interface{} `bson:"id"`
	Result interface{} `bson:"result,omitempty"`
	Error *bsonError `bson:"error,omitempty"`
}


// processBsonRpc processes a json rpc request BSON encoded, reading 
// from the io.Reader and sending the
// result to the io.Writer.
func processBsonRpc(in io.Reader, out io.Writer, context *_root.{{.ClassName}}) {
	var buf bytes.Buffer
	io.Copy(&buf, in)

	glog.Infof("processBsonRpc...: %s", buf.String())
	var request jsonRequest

	if err := bson.Decode(buf.Bytes(), &request); nil!=err {
		errorBsonRpc(out, request.Id, BSONRPC_ERROR_PARSE_ERROR, err, nil)
		return
	}
	switch request.Method {
	{{range .Methods}}
	{{if .Abbreviation}}
	case "{{.Abbreviation}}":
		fallthrough
	{{end}}
	case "{{.Name}}":
		// args := struct {
		// 	{{range $i,$p := .Parameters}}
		// 	{{$p.Field.TitleName}} {{$p.Field.GoType "_root"}} `bson:"{{$i}}"`
		// 	{{end}}
		// }{}
		// if err = request.Params.Unmarshal(&args); nil!=err {
		// 	errorBsonRpc(out, request.Id, BSONRPC_ERROR_INVALID_PARAMS, err, nil)
		// 	return
		// }
		
		{{if .Results.Length}}
		result := make([]interface{}, {{.Results.Length}})
		{{else}}
		result := []interface{}{}
		{{end}}

		if err = func() (err error) {
			{{/*
				We catch any panic inside the method itself
				and convert it into an error return
			*/}}
			defer func() {
				if r:=recover(); nil!=r {
					if e, ok := r.(error); ok {
						err=e
					} else {
						err = fmt.Errorf("PANIC: %s", e)
					}
				}
			}()

			{{if .Results.Length}}
				{{range .Results.LengthArray}}{{if .}},{{end}}result[{{.}}]{{end}} = {{template "execute" .}}
				{{if .Results.IsErrorLast}}
				if (nil!=result[{{.Results.LastElementIndex}}]) {
					return result[{{.Results.LastElementIndex}}].(error)
				}
				{{/*
				 If the method ends in an error, and we don't have
				 an error after calling the method, we remove the error :
				 no need to return it to the caller.
				*/}}
				result = result[0:{{.Results.LastElementIndex}}]
				{{end}}{{/* Of No last error in result */}}

			{{else}}
			{{/*
			 If the method returns no result, we still execute
			 it in the web server thread since it might panic
			 */}}
			{{template "execute" .}}
			{{end}}
			return nil
		}(); nil!=err {
			errorBsonRpc(out, request.Id, BSONRPC_ERROR_APPLICATION_ERROR, err, nil)
			return
		}

		{{/*
			Our result is an []interface{}
			with all the values we want to return
		*/}}
		response := bsonResponse{
			Jsonrpc:"2.0",
			Id: request.Id,
			Result: result,
		}
		bsonRaw, err := bson.Marshal(&response)
		if nil!=err {
			errorBsonRpc(out, request.Id, BSONRPC_ERROR_INTERNAL_ERROR, err, nil)
			return
		}
		_, err := out.Write(bsonRaw)
		if nil!=err {
			errorBsonRpc(out, request.Id, BSONRPC_ERROR_INTERNAL_ERROR, err, nil)
		}

	{{end}} {{/* End of range across methods */}}
	default:
		errorBsonRpc(out, request.Id, BSONRPC_ERROR_METHOD_NOT_FOUND, 
			errors.New("Method " + request.Method + " not found"),
			 nil)
		return
	}
}


// errorBsonRpc sends a JSONRpcError back to the client.
func errorBsonRpc(out io.Writer, id interface{}, code int, err error, data interface{}) {
	glog.Infof("ERROR on BsonRPC: %d, %s {%q}", code, err, data)
	response := bsonResponse{
		Jsonrpc:"2.0",
		Id: id,
		Error: &bsonError {
			Code:code,
			Message: err.Error(),
			Data: data,
		},
	}

	bsonRaw, encErr := bson.Marshal(&response)
	// If our encoding fails on error response, we assume it's because
	// the data encoding failed, so we try again, but without the data,
	// instead sending our failed error message as the data
	if nil!=encErr && nil!=data {
		errorBsonRpc(out, id, code, err, encErr.Error())
		return
	}
	if _, err := out.Write(bsonRaw); nil!=err {
		errorBsonRpc(out, id, code, err, nil)
		return
	}
}
