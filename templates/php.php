<?php
/**
 * {{.PackageName}} is the golait2-generated class
 * for doing JSONRpc calls to the golait2 based server.
 */
class {{.PackageName}}_Result {
	var $Error;
	var $Result;

	public function __construct($error, $results=null) {
		$this->Error = $error;
		$this->Result = $results;
	}
}

class {{.PackageName}} {
	var $_jsonUrl;
	protected static $_id=1;

	public function __construct($server, $jsonPath="/rpc/{{.ClassName}}/json") {
		$this->_jsonUrl = $server . $jsonPath;
	}
	{{range .Methods}}
	public function {{.Name}} ({{.ParameterNameList | Prefix "$" | Join ","}}) {
		$args = array(
		{{- range .Parameters}}
			{{.Field.CoerceInPHP}},
		{{- end}}
		);

		return $this->_rpc("{{.Name}}", $args);
	}
	{{end}}

	protected function _rpc($method, $argsArray) {
		$id = self::$_id++;

		$request = array(
			"id"=>"$id",
			"method"=>$method,
			"params"=>$argsArray
		);
		$requestJson = json_encode($request);
		error_log(__FILE__ . ":" . __LINE__ . ": URL = " . $this->_jsonUrl . ", Request = " . $requestJson);
		$c = curl_init($this->_jsonUrl);
		curl_setopt($c, CURLOPT_POST, 1);
		curl_setopt($c, CURLOPT_POSTFIELDS, $requestJson);
		curl_setopt($c, CURLOPT_RETURNTRANSFER, TRUE);
		$res = curl_exec($c);
		if (FALSE===$res) {
			$err = curl_error($c);
			curl_close($c);
			return new {{.PackageName}}_Result("Request " . $requestJson . " to $this->_jsonUrl failed: $c");
		}
		curl_close($c);
		$json = json_decode($res);
		if (JSON_ERROR_NONE!=json_last_error()) {
			error_log(__FILE__ . ":" . __LINE__ . ": FAILED TO DECODE JSON $res: " . json_last_error_msg());
			return new {{.PackageName}}_Result(json_last_error_msg(), FALSE);
		}
		$err = false;
		if (property_exists($json, "error") && $json->error) {
			$err = $json->error->message;
			$result = property_exists($json, "result") ? $json->result : false;
			return new {{.PackageName}}_Result($err, $result);
		}
		if (!property_exists($json, "result")) {
			error_log(__FILE__ . ":" . __LINE__ . ": FAILED TO GET A result VALUE for $method");
			return new {{.PackageName}}_Result("An error does not appear to have occurred but the request failed to provide a result value.", false);
		}
		return new {{.PackageName}}_Result($err, $json->result);
	}
}
