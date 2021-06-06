package {{.PackageName}}

import (
	"bytes"
	"io"
	"net/http"
	"log"

	"github.com/gorilla/websocket"
	"github.com/golang/glog"

	_root "{{.RootPackagePath}}"
)

var bson_upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func WsBsonHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Ensure that a write to a closed channel
	// won't panic us out of existence	
	defer recover()

	glog.Infof("Got a bson connection")
	defer glog.Infof("BSON Connection closed")

	{{if .Constructor}}
	context := _root.{{.Constructor}}(nil,r)
	{{else}}
	context := &_root.{{.ClassName}}{}
	{{end}}

	conn, err := bson_upgrader.Upgrade(w,r, nil)
	if err!=nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()


	for {
		typ, in, err := conn.NextReader()
		if nil!=err {
			log.Printf("ERROR obtaining WS NextReader: %s", err.Error())
			break
		}
		switch typ {
		case websocket.TextMessage:
			fallthrough
		case websocket.BinaryMessage:
			// go func () {
			// 	var buf bytes.Buffer
			// 	processJsonRpc(in, &buf, context)
			// 	OUT <- buf.Bytes()
			// }()
			/*
			out,err := conn.NextWriter(websocket.TextMessage)
			if nil!=err {
				log.Printf("ERROR obtaining WS NextWriter: %s", err.Error())
				return
			}
			// Because of how gorilla websocket NextReader and NextWriter
			// work, we can't run this in a goroutine, which seems a pity...

			processJsonRpc(in, out, context)
			out.Close()
			*/
			var buf bytes.Buffer
			io.Copy(&buf, in)
			go func() {
				var outbuf bytes.Buffer
				processBsonRpc(bytes.NewReader(buf.Bytes()), &outbuf, context)
				conn.WriteMessage(websocket.BinaryMessage, outbuf.Bytes())
			}()
			
		case websocket.CloseMessage:
			return	// Exit from our infinite loop
		case websocket.PingMessage:
			// Let the default handler send the Pong
		case websocket.PongMessage:
			// Let the default handler do nothing
		} 
	}
}