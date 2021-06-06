
class {{.ClassName}}
	constructor: (path="/rpc/{{.ClassName}}/json", server="", @_timeout=0)->
		if ""==server
			server = document.location.protocol + "//" + document.location.host
		@url = server + path
		@ClearRPCErrorHandler()
	{{range .Methods}}
	{{.Name}}: ({{.ParameterNameList | Join ","}})->		
		@_rpc("{{.Name}}",  arguments )
	{{end}}
	SetTimeout: (ts)->
		@_timeout = ts
	SetRPCErrorHandler: (@_errorHandler=null)->
		if not @_errorHandler?
			@_errorHandler = (e)->
				console.error?("RPC Error: ", e)
				alert("RPC ERROR: " + e)
				return
		return
	ClearRPCErrorHandler: ()->
		@_errorHandler = false
		return
	_reject: (reject,err)->
		if @_errorHandler
			@_errorHandler(err)
			return
		reject(err)
		return
	_rpc: (method, params)->
		# params comes in as a 'arguments' object, so we need to
		# convert to an array
		params = (params[i] for i in [0...params.length])
		# console.log("params = ", params)
		# if 1==params.length
		# 	params = [params[0]]
		# else
		# 	params = Array.apply(null, params)

		new Promise( (resolve, reject)=>
			if (window.XMLHttpRequest)
				req = new XMLHttpRequest()
			else if (window.ActiveXObject)
				req = new ActiveXObject("Microsoft.XMLHTTP")
			else
				@_reject(reject, "No supported HttpRequest implementation")
				return

			bind = (resolve, reject, req)=>
				()=>
					if 4==req.readyState 
						if 200==req.status
							res = req.response
							if ! res?
								@_reject(reject, "Failed to parse response: " + req.response)
								return
							if res.error?
								@_reject(reject, res.error)
								return
							if res.result?
								resolve(res.result)
								return
							# This is a send-and-forget JSON RPC request (ie one without id)
							resolve(null)
							return
						@_reject(reject, "Failed with " + req.statusText)

			req.onreadystatechange = bind(resolve, reject, req)
			req.timeout = @_timeout
			req.open("POST", @url + "?" + method, true)
			req.responseType = "json"
			req.send(JSON.stringify({ id: @_id++, method:method, params: params }))
		)
	_id: 0

# Make our class RequireJS/AMD Compatible
window.{{.ClassName}} = {{.ClassName}}

if typeof window.define is "function" && window.define.amd
	window.define("{{.ClassName}}", [], () -> window.{{.ClassName}})
