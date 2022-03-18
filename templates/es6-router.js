// Transport is both transport (http/ws) and protocol (jsonrpc/other)
// agnostic. It just handles a sending (with queueing in the event that
// a transport takes a while to start) and receiving.
// Anyone can send with .send(data), and anyone who registers as a 
// addListener() will receive transport_opened(), transport_closed() and
// transport_received() callback, unless an earlier listener returns 'true' 
// for transport_received()
{{if .JsExport}}export {{end -}}
class {{.ClassName}}Transport {
	constructor(baseTransport) {
		this.listeners = [];
		this._open = false;
		this._id = 0;
		this.base = baseTransport;
		this.base.addListener(this);
	}
	addListener(listener) {
		this.listeners.push(listener);
		if (this.base.isOpen()) {
			listener.transport_opened(this);
		}
	}
	removeListener(l) {
		this.listeners = this.listeners.filter( m=>l!=m);
	}
	// Called by the implementation when the transport is open and sends will succeed
	opened(t) {
		this.listeners.map( l=>l.transport_opened(this) );
	}
	closed(t) {
		this.listeners.map( l=>l.transport_closed(this) );
	}
	send(data) {
		let req = {
			data: data,
			resolve: null,
			reject: null,
		}
		return new Promise( (resolve,reject)=>{
			req.resolve = resolve;
			req.reject = reject;
			this.base.send(req);
		})
	}
	// received calls each listeners, until one returns true for transport_received(data)
	received(data) {
		for (let l of this.listeners) {
			if (l.transport_received(this, data)) break;
		}
	}
	error(t, err) {
		this.listeners.map(l=>l.transport_error(this,err));
	}
}


// TransportWs uses a Websocket to send/receive data for a generic
// transport
{{if .JsExport}}export {{end -}}
class {{.ClassName}}TransportWs {
	constructor(wsfactory) {
		this.wsfactory = wsfactory;
		this.listeners = [];
		this.queue = [];
		this.promises = {};
		this._id = 0;
		this.createWs();
	}
	createWs() {
		this.ws = this.wsfactory();
		this.ws.addEventListener(`open`, evt=>{
			this.queue.map( req=>this.send(req) );
			this.queue = [];			
			this.listeners.map( l=>{
				if (!l.opened) debugger;
				l.opened(this);
			} );
		});
		this.ws.addEventListener(`message`, evt=>{
			// console.log(`{{.ClassName}}TransportWs.message, evt=`, evt);
			let js = JSON.parse(evt.data);
			// if we've received valid a valid jsonrpc result | method call
			if (js) {
				if (js.error || js.result) {
					let req = this.promises[js.id];
					if (req) {
						if (undefined != js.error) {
							req.reject(js.error);
						} else {
							req.resolve(js.result);
						}
						delete this.promises[js.id];	// remove this promise from our list
						// we've dealt with this message so no reason to pass it on
						return;
					}
				} else if (js.method) {
					// console.log(`Received jsonrpc call from ws: `, js);
				}
				this.listeners.map(l=>l.received(js));
				return;
			}
			this.listeners.map(l=>l.received(evt.data));
		});
		this.ws.addEventListener(`close`, evt=>{
			this.listeners.map(l=>{
				if (!l.closed) debugger;
				l.closed(this);
			});
			this.ws = null;
			setTimeout(()=>this.createWs(), 500); // re-establish conn in 0.5s
		});
		this.ws.addEventListener(`error`, evt=>{
			this.listeners.map(l=>l.error(this, evt));
		});
	}
	addListener(l) {
		// console.log(`{{.ClassName}}TransportWs::addListener`, l);
		this.listeners.push(l);
	}
	removeListener(l) {
		this.listeners = this.listeners.filter( m=>l!=m );
	}
	send(req) {
		let id = this._id++;
		if (!req.data[`id`]) req.data[`id`] = id;
		if (!req.data[`jsonrpc`]) req.data[`jsonrpc`] = `2.0`;
		if (this.isOpen()) {
			this.ws.send(JSON.stringify(req.data));
		} else {
			this.queue.push(req);
		}
		this.promises[req.data[`id`]] = req;
	}
	isOpen() {
		return (this.ws) && (1==this.ws.readyState);
	}
}

// {{.ClassName}}TransportHttp does http(s) transport
// for messaging. It sends responses immediately.
{{if .JsExport}}export {{end -}}
class {{.ClassName}}TransportHttp {
	// fetchf should be something like
	// 			fetch(this.url + "?" + method, {
	//				method: 'POST',
	//			headers: this.headers(),
	//			body: JSON.stringify({ id: this._id++, method:method, params: params })
	//		})
	constructor(fetchf) {
		this.fetchf =fetchf;
		this.listeners = [];

		// path="/rpc/{{.ClassName}}/json", server="", _timeout=0) {
		// this._id = 0;
		// this._timeout = 0;
		// if (""==server) {
		// 	server = document.location.protocol + "//" + document.location.host;
		// }
		// this.url = server + path;
		// this._headers = {
		// 	"Accept": "application/json",
		// 	"Content-Type":"application/json"
		// };
	}
	addListener(l) {
		this.listeners.push(l);
	}
	removeListener(l) {
		this.listeners = this.listeners.filter(m=>l!=m);
	}
	// send directly calls the resolve/reject of the
	// request
	send(req) {
		// ID is not important with send since we get the answer
		// at once.
		this.fetchf(req.data)
		.then( res=>{
			if (!res.ok) {
				let err = new Error(res.statusText);
				err.response = res;
				throw err
			}
			req.resolve(res);
		}).catch( err=>req.reject(err) );
	}
	isOpen() {
		return true;
	}
}


/**
 * {{.ClassName}}Router provides the API
 * in combination with a transport. The transport can be
 * http or ws. The router handles the method call and responses,
 * while the transport is responsible for delivery and receipt to the server.
 */
{{if .JsExport}}export {{end -}}
class {{.ClassName}}Router {
	constructor(transport, clientApi=null) {
		this.transport = new {{.ClassName}}Transport(transport);
		this._id = 0;
		this._errorHandler = null;
		this.setErrorHandler(null);
		this.transport.addListener(this);
		this.clientApi = clientApi;
		this.listeners = [];
	}
	addListener(l) {
		this.listeners.push(l);
	}
	removeListener(l) {
		this.listeners = this.listeners.filter(m=>l!=m);
	}
	transport_opened(t) {
		// console.log(`{{.ClassName}}Router::transport_opened`);
		this.listeners.map( l=> {
			if (l.transport_opened) l.transport_opened(this,t);
		});
	}
	transport_closed(t) {
		// console.log(`{{.ClassName}}Router::transport_closed`);
		this.listeners.map( l=> {
			if (l.transport_closed) l.transport_closed(this,t);
		});
	}
	transport_received(t, msg) {
		// console.log(`{{.ClassName}}Router::transport_received`, msg);
		if ('object'==typeof msg && 'undefined'!=typeof msg.method) {
			let p = this.clientApi[msg.method].apply(this.clientApi, msg.params);
			// @TODO If we have an ID on the message, we should send a response
			if (msg.id) {
				p.then( res=>{
					this.transport.send({id:msg.id, response:res});
				}).catch( err=> {
					this.transport.send({id:msg.id, error: err});
				});
			}
		}
		this.listeners.map( l=> {
			if (l.transport_received) l.transport_received(this,t,msg);
		});
	}
	transport_error(t, err) {
		// console.error(`{{.ClassName}}Router::transport_error`, err);
		this.listeners.map( l=> {
			if (l.transport_error) l.transport_error(this,t,err);
		});
	}
	setErrorHandler(_errorHandler=null) {
		this._errorHandler = _errorHandler;
	}
	_reject (reject,err) {
		if (null!=this._errorHandler) {
			this._errorHandler(err);
		}
		reject(err);
		return;
	}
	_rpc(method, params) {
		let id = this._id++;
		// params comes in as an 'arguments' object, so we need to convert
		// it to an array
		params = Array.prototype.slice.call(params);
		return this.transport.send({ id:id, method:method, params:params});
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
