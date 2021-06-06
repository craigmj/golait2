
class {{.ClassName}}Http {
	constructor(path="/rpc/{{.ClassName}}/json", server="", _timeout=0) {
		this._id = 0;
		this._timeout = 0;
		if (""==server) {
			server = document.location.protocol + "//" + document.location.host;
		}
		this.url = server + path;
		this._headers = {
			"Accept": "application/json",
			"Content-Type":"application/json"
		};
	}
	setTimeout (ts) {
		this._timeout = ts;
	}
	reject(reject, err) {
		reject(err);
	}
	headers() {
		return this._headers;
	}
	rpc (method, params) {
		// params comes in as an 'arguments' object, so we need to convert
		// it to an array
		params = Array.prototype.slice.call(params);

		return new Promise( (resolve, reject)=>{		
			fetch(this.url + "?" + method, {
				method: 'POST',
				headers: this.headers(),
				body: JSON.stringify({ id: this._id++, method:method, params: params })
			})
			.then( (res)=>{
				if (!res.ok) {
					let err = new Error(res.statusText);
					err.response = res;
					throw err;
				}
				return res.json();
			}).then( (js)=>{
				if (null!=js.error) {
					throw js.error;
				}
				if (null!=js.result) {
					resolve(js.result);
					return;
				}
				resolve(null);
				return;
			} )
			.catch( (err)=>{
				this.reject(reject, err);
			})
		});
	}
	{{range .Methods}}
	{{.Name}}({{.ParameterNameList | Join ","}}) {
		return this.rpc("{{.Name}}",  arguments );
	}
	{{end}}
	flatten(callback, context=null) {
		return function(argsArray) {
			callback.apply(context, argsArray);
		}
	}
}

// Define the class in the window and make AMD compatible
window.{{.ClassName}} = {{.ClassName}}Http;
if (("function" == typeof window.define) && (window.define.amd)) {
	window.define("{{.ClassName}}", [], function() { return window.{{.ClassName}}; });
}
