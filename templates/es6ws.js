class {{.ClassName}}Ws {
	constructor(path="/rpc/{{.ClassName}}/json/ws", server="") {
		if (""== server) {
			server = "ws" +
				 ("https:" == document.location.protocol ? "s" : "") +
				 "://" +
				 document.location.host
		}
		this._id = 0;
		this.url = server+path;
		this.live = {};
		this.queue = [];
		this.setRPCErrorHandler(null);
		this.startWs();
	}	
	setRPCErrorHandler(_errorHandler=null) {
		this._errorHandler = _errorHandler;
	}
	_reject (reject,err) {
		if (null!=this._errorHandler) {
			this._errorHandler(err);
		}
		reject(err);
		return;
	}
	startWs() {
		this.ws = new WebSocket(this.url);
		this.ws.onmessage = (evt)=> {
			let log = console.error ? console.error : console.log;
			let res = false;
			try {
				res = JSON.parse(evt.data);
			} catch (err) {
				log(`Failed to parse JSON response:`, evt.data);
				return;
			}
			if (undefined == res || undefined==res.id) {
				log(`Failed to parse response: ${evt.data}`);
				return;
			}
			let promise = this.live[res.id];
			if (undefined == promise) {
				log(`Failed to find promise for ${evt.data}`);
				return;
			}
			delete this.live[res.id];
			if (null!=res.error) {
				this._reject(promise.reject, res.error);
				return;
			}
			if (null!=res.result) {
				promise.resolve(res.result);
				return;
			}
			promise.resolve(null);
		};
		this.ws.onerror = (err)=> {
			console.error("ERROR on websocket:", err);
		};
		this.ws.onopen = (evt)=> {
			console.log("Connected websocket");
			for (let q of this.queue) {
				this.ws.send(q);
			}
			console.log(`Emptied queue of ${this.queue.length} queued messages`);
			this.queue = [];
		};
		this.ws.onclose = (evt)=> {
			console.log("Websocket closed - attempting reconnect in 1s");
			setTimeout( ()=> this.startWs(), 1000 );
		}
	}
	_rpc(method, params) {
		let id = this._id++;
		// params comes in as an 'arguments' object, so we need to convert
		// it to an array
		params = Array.prototype.slice.call(params);
		// let p = [];
		// for (let i=0; i<p.length; i++) {
		// 	p[i] = params[i]
		// }

		let data = JSON.stringify({ id:id, method:method, params:params })
		this.live[id] = {
			resolve: null,
			reject: null
		};
		return new Promise( (resolve,reject)=> {
			this.live[id].resolve = resolve;
			this.live[id].reject = reject;
			if (1==this.ws.readyState) {
				this.ws.send(data);
			} else {
				this.queue.push(data);
			}
		});
	}

	{{range .Methods}}
	{{.Name}} ({{.ParameterNameList | Join ","}}) {
		return this._rpc("{{.Name}}",  arguments );
	}
	{{end}}

	flatten(callback, context=null) {
		return function(argsArray) {
			callback.apply(context, argsArray);
		}
	}
}

// Define the class in the window and make AMD compatible
window.{{.ClassName}}Ws = {{.ClassName}}Ws;
if (("function" == typeof window.define) && (window.define.amd)) {
	window.define("{{.ClassName}}Ws", [], function() { return window.{{.ClassName}}Ws; });
}
