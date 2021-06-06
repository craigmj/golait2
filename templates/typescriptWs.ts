enum WSState {
	Null = 0,
	Connecting = 1,
	Connected = 2
}


export class {{.ClassName}}Ws {
	protected id: number;
	protected url: string;
	protected live:Map<number,[(...a:any[])=>void,(a:any)=>void]>;
	protected queue: string[];
	protected errorHandler: (a:any)=>void;
	protected ws: WebSocket;
	protected wsState: WSState;

	constructor(path:string="/rpc/{{.ClassName}}/json/ws", server:string="") {
		if (""== server) {
			server = "ws" +
				 ("https:" == document.location.protocol ? "s" : "") +
				 "://" +
				 document.location.host
		}
		this.id = 0;
		this.url = server+path;
		this.live = new Map<number,[(...a:any[])=>void,(a:any)=>void]>();
		this.queue = new Array<string>();
		this.setRPCErrorHandler(null);
		this.wsState = WSState.Null;
		this.startWs();
	}	
	setRPCErrorHandler(handler?:(a:any)=>void) :void {
		this.errorHandler = handler;
	}
	reject(reject:(a:any)=>void,err:Error):void {
		if (this.errorHandler) {
			this.errorHandler(err);
		}
		reject(err);
		return;
	}
	startWs():void {
		if (this.wsState!=WSState.Null) {
			return;
		}
		this.wsState = WSState.Connecting;

		this.ws = new WebSocket(this.url);
		this.ws.onmessage = (evt)=> {
			let res:any = JSON.parse(evt.data);
			if (undefined == res || undefined==res.id) {
				console.error(`Failed to parse response: ${evt.data}`);
				return;
			}
			let id = res.id as number;
			let promises = this.live.get(id);
			if (! promises) {
				console.error(`Failed to find promise for ${evt.data}`);
				return;
			}
			this.live.delete(id);
			let [resolve,reject] = promises;

			if (res.error) {
				this.reject(reject, res.error);
				return;
			}
			if (res.result) {
				resolve(res.result);
				return;
			}
			resolve(undefined);
		};
		this.ws.onerror = (err)=> {
			console.error("ERROR on websocket:", err);
			this.wsState = WSState.Null;
		};
		this.ws.onopen = (evt)=> {
			this.wsState = WSState.Connected;
			console.log("Connected websocket");
			for (let q of this.queue) {
				this.ws.send(q);
			}
			console.log(`Emptied queue of ${this.queue.length} queued messages`);
			this.queue = [];
		};
		this.ws.onclose = (evt)=> {
			console.log("Websocket closed - attempting reconnect in 1s");
			this.wsState = WSState.Null;
			setTimeout( ()=> this.startWs(), 1000 );
		}
	}
	rpc(method:string, params:any[]) {
		let id = this.id++;
		// // params comes in as an 'arguments' object, so we need to convert
		// // it to an array
		// params = Array.prototype.slice.call(params);
		// // let p = [];
		// // for (let i=0; i<p.length; i++) {
		// // 	p[i] = params[i]
		// // }

		let data = JSON.stringify({ id:id, method:method, params:params });
		this.live.set(id, [undefined,undefined]);
		return new Promise( (resolve:(...a:any[])=>void,reject:(a:any)=>void)=> {
			if (this.wsState==WSState.Null) {
				this.startWs();
			}
			this.live.set(id,[resolve,reject]);
			if ((this.wsState == WSState.Connected) && (1==this.ws.readyState)) {
				this.ws.send(data);
			} else {
				this.queue.push(data);
			}
		});
	}

	{{range .Methods}}
	{{.Name}} ({{.ParameterTypedNameListTs | Join ","}}) {
		return this.rpc("{{.Name}}",  [{{.ParameterNameList | Join ","}}] );
	}
	{{end}}
}
