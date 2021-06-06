class {{.ClassName}}Ws
	constructor: (path="/rpc/{{.ClassName}}/json/ws", server="")->
		if ""== server
			server = "ws" +
				 (if "https" == document.location.protocol then "s" else "") +
				 "://" +
				 document.location.host
		@url = server + path
		@live = {}
		@queue = []
		@SetRPCErrorHandler(null)
		@startWs()
		
	SetRPCErrorHandler: (@_errorHandler=null)->
		if not @_errorHandler?
			@_errorHandler = (e)->
				console.error?("RPC Error: ", e)
				alert("RPC ERROR: " + e)
				return
		return
	_reject: (reject,err)->
		if @_errorHandler
			@_errorHandler(err)
			return
		reject(err)
		return

	{{range .Methods}}
	{{.Name}}: ({{.ParameterNameList | Join ","}})->
		@_rpc("{{.Name}}",  arguments )
	{{end}}

	startWs: ()->
		@ws = new WebSocket(@url)
		@ws.onmessage = (evt)=>
			res = JSON.parse(evt.data)
			if !(res? && res.id?)
				console.error?("Failed to parse response: #{evt.data}")
				return
			promise = @live[res.id]
			if !promise?
				console.error?("Failed to find promise for: #{evt.data}")
				return
			delete @live[res.id]
			if res.error?
				@_reject(promise.reject, res.error)
				return
			if res.result?
				promise.resolve(res.result)
				return
			promise.resolve(null)
		@ws.onerror = (err)->
			console.error?("ERROR on websocket:", err)
		@ws.onopen = (evt)=>
			console.log("Connected websocket")
			for q in @queue
				@ws.send(q)
			@queue = []
		@ws.onclose = (evt)=>
			console.log?("Websocket closed - attempting reconnect in 2s")
			# setTimeout(()=>
			# 	@startWs()
			# , 2000)
	_rpc: (method, params)->
		id= @_id++
		# params comes in as a 'arguments' object, so we need to
		# convert to an array
		params = (params[i] for i in [0...params.length])

		data = JSON.stringify({id:id, method:method, params:params })
		# console.log("Sending ", data)
		@live[id] = {
			resolve: null,
			reject: null
		}
		new Promise( (resolve,reject)=>
			@live[id].resolve = resolve
			@live[id].reject = reject
			if 1==@ws.readyState
				@ws.send(data)
			else
				@queue.push(data)
		)
	_id: 0


# Make our class RequireJS/AMD Compatible
window.{{.ClassName}}Ws = {{.ClassName}}Ws

if typeof window.define is "function" && window.define.amd
	window.define("{{.ClassName}}Ws", [], () -> window.{{.ClassName}}Ws)
