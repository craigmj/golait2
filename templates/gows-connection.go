package {{.PackageName}}

{{$ConnectionClass := .ConnectionClass}}
{{$ConnectionClassConstructor := .ConnectionClassConstructor}}

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	_root "{{.RootPackagePath}}"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func WsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Ensure that a write to a closed channel
	// won't panic us out of existence	
	defer recover()

	if err := _root.{{$ConnectionClassConstructor}}(nil, r, func(conn *_root.{{$ConnectionClass}})error {
		wsConn, err := upgrader.Upgrade(w,r, nil)
		if err!=nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		defer wsConn.Close()

		connLock := &sync.Mutex{}

		// Trigger a Ping every 45s to ensure our connection remains alive
		stopPing := make(chan bool)
		defer func() {
			stopPing <- true
		}()
		go func() {
			ticker := time.NewTicker(45 * time.Second)
			for {
				select {
				case <-ticker.C:
					connLock.Lock()
					wsConn.WriteMessage(websocket.PingMessage, []byte("ai"))
					connLock.Unlock()
				case <-stopPing:
					ticker.Stop()
					return
				}
			}
		}()

		for {
			typ, in, err := wsConn.NextReader()
			if nil!=err {
				Log.Errorf("ERROR obtaining WS NextReader: %s", err.Error())
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
				out,err := wsConn.NextWriter(websocket.TextMessage)
				if nil!=err {
					Log.Printf("ERROR obtaining WS NextWriter: %s", err.Error())
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
					processJsonRpc(bytes.NewReader(buf.Bytes()), &outbuf, conn)
					connLock.Lock()
					wsConn.WriteMessage(websocket.TextMessage, outbuf.Bytes())
					connLock.Unlock()
				}()
				
			case websocket.CloseMessage:
				return nil	// Exit from our infinite loop
			case websocket.PingMessage:
				// We never get here - see SetPingHandler : Let the default handler send the Pong
			case websocket.PongMessage:
				// We never get here - set SetPongHandler : Let the default handler do nothing
			} 
		}
		return nil
	});nil!=err {
		Log.Errorf("Error occurred on WS: %s", err.Error())
		return
	}
}