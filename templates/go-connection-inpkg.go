package {{.PackageName}}
{{$packageName := .PackageName -}}
{{$Recover := .Recover -}}

{{$ConnectionClass := .ConnectionClass -}}
{{$ConnectionClassConstructor := .ConnectionClassConstructor -}}

{{/*
The 'execute' template should be called on a MethodDefinition
object, and will write out the code to execute the method itself.
*/ -}}
{{define "execute" -}}
context.{{.Name}}(
		{{range $i, $p := .Parameters -}}
			args.{{$p.Field.TitleName}},
		{{end }}
				)
{{end}}

import (
{{with .Imports -}}
	{{.AddImports "errors" "fmt" "io" "encoding/json" "net/http" "bytes" "_golog log"}}
	{{range .Imports }}{{if .ShouldImport -}}
	{{.Name}} {{.Path}}{{end}}
	{{end -}}
{{end}}
)

// Just a placeholder to prevent errors if fmt isn't used
var _ = fmt.Println

/** Logger is a simple logging interface that will by default write to log.
 * To provide your own logger, set the Log class to an instance implementing
 * the Logger interface.
 */
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type logger struct {}
func (l *logger) Infof(format string, args ...interface{}) { _golog.Printf(format, args...)}
func (l *logger) Errorf(format string, args ...interface{}) { _golog.Printf("ERROR: "+format, args...)}

var Log Logger
func init() {
	Log = &logger{}
}

/**
 * HttpHandlerFunc is the entry point for the http POST
 * handler.
 */
func HttpHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if err := {{$ConnectionClassConstructor}}(w,r,func(conn *{{$ConnectionClass}}) error {
		return ProcessJsonRpc(r.Body, w, conn)
	}); nil!=err {
		Log.Errorf("Error occurred: %s", err.Error())
		http.Error(w,err.Error(), http.StatusInternalServerError)
		return
	}
}

const JSONRPC_ERROR_PARSE_ERROR = -32700
const JSONRPC_ERROR_INVALID_REQUEST = -32600
const JSONRPC_ERROR_METHOD_NOT_FOUND = -32601
const JSONRPC_ERROR_INVALID_PARAMS = -32602
const JSONRPC_ERROR_INTERNAL_ERROR = -32603
const JSONRPC_ERROR_APPLICATION_ERROR = -1000


type jsonRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Method string `json:"method"`
	Params []json.RawMessage `json:"params"`
	Id	interface{} `json:"id"`
}
type jsonError struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data interface{} `json:"data,omitempty"`
}
type jsonResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id interface{} `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error *jsonError `json:"error,omitempty"`
}

// ProcessJsonRpc processes a json rpc request, reading 
// from the io.Reader and sending the
// result to the io.Writer.
func ProcessJsonRpc(in io.Reader, out io.Writer, conn *{{$ConnectionClass}}) error {
	var buf bytes.Buffer
	var err error
	io.Copy(&buf, in)

	Log.Infof("ProcessJsonRpc: %s", buf.String())
	var request jsonRequest

	js := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err = js.Decode(&request); nil!=err {
		errorJsonRpc(out, request.Id, JSONRPC_ERROR_PARSE_ERROR, err, nil)
		return err
	}
	if err = conn.Context(func (context *{{.ClassName}})error {
		switch request.Method {
		{{range .Methods}}
		{{if .Abbreviation}}
		case "{{.Abbreviation}}":
			fallthrough
		{{end}}
		case "{{.Name}}":
			{{/*
			I only really need to consider a few types
			Numbers, Strings, booleans, structs
			since JSON doesn't support that many types
			itself.
			*/}}
			{{if .Parameters}}
			args := struct {
				{{range $i,$p := .Parameters}}
				{{$p.Field.TitleName}} {{$p.Field.GoType ""}} `json:"{{$i}}"`
				{{end}}
			}{}

			// Decoding request.Params as an array
			if len(request.Params)!={{.ParametersLen}} {
				err = fmt.Errorf(
					"Expected %d parameters, but got %d in call to %s", {{.ParametersLen}}, len(request.Params), "{{.Name}}")
				errorJsonRpc(out, request.Id, JSONRPC_ERROR_INVALID_PARAMS, err, nil)
				return err
			}
			{{range $i, $p := .Parameters}}
			if err = json.Unmarshal(request.Params[{{$i}}], &args.{{$p.Field.TitleName}});
				nil!=err {
				errorJsonRpc(out, request.Id, JSONRPC_ERROR_INVALID_PARAMS, fmt.Errorf(
					"Unable to decode JSON param %d: %s", {{$i}}+1, err.Error()), nil)
				return err
			}
			{{end}}
			{{end}}

			// // Decoding request.Params as an object
			// err := json.Unmarshal(request.Params, &args)
			// if nil!=err {
			// 	errorJsonRpc(out, request.Id, JSONRPC_ERROR_INVALID_PARAMS, err, nil)
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
				{{if $Recover}}
				defer func() {
					if r:=recover(); nil!=r {
						if e, ok := r.(error); ok {
							err=e
						} else {
							err = fmt.Errorf("PANIC: %s", e)
						}
					}
				}()
				{{end}}

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
				errorJsonRpc(out, request.Id, JSONRPC_ERROR_APPLICATION_ERROR, err, nil)
				return err
			}

			{{/*
				Our result is an []interface{}
				with all the values we want to return
			*/}}
			response := jsonResponse{
				Jsonrpc:"2.0",
				Id: request.Id,
				Result: result,
			}
			encoder := json.NewEncoder(out)
			if err := encoder.Encode(&response); nil!=err {
				errorJsonRpc(out, request.Id, JSONRPC_ERROR_INTERNAL_ERROR, err, nil)
				return err
			}

		{{end}} {{/* End of range across methods */}}
		default:
			errorJsonRpc(out, request.Id, JSONRPC_ERROR_METHOD_NOT_FOUND, 
				errors.New("Method " + request.Method + " not found"),
				 nil)
			return err
		}		
		return nil
	}); nil!=err {
		errorJsonRpc(out, request.Id, JSONRPC_ERROR_APPLICATION_ERROR, err, nil)
		return err
	}
	return nil
}


// errorJsonRpc sends a JSONRpcError back to the client.
func errorJsonRpc(out io.Writer, id interface{}, code int, err error, data interface{}) {
	Log.Infof("ERROR on JsonRPC: %d, %s {%q}", code, err, data)
	response := jsonResponse{
		Jsonrpc:"2.0",
		Id: id,
		Error: &jsonError {
			Code:code,
			Message: err.Error(),
			Data: data,
		},
	}
	js := json.NewEncoder(out)
	// If our encoding fails on error response, we assume it's because
	// the data encoding failed, so we try again, but without the data,
	// instead sending our failed error message as the data
	if encErr := js.Encode(&response); nil!=encErr && nil!=data {
		errorJsonRpc(out, id, code, err, encErr.Error())
	}
}
