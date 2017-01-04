package birect

import "github.com/marcuswestin/go-ws"

// Client is used register request handlers (for requests sent from the server),
// and to send requests to the server.
type Client struct {
	jsonReqHandlerMap
	protoReqHandlerMap
	*Conn

	// Temporary
	OnDisconnectHack func()
}

// Connect connects to a birect server at address
func Connect(address string) (client *Client, err error) {
	client = &Client{
		jsonReqHandlerMap:  make(jsonReqHandlerMap),
		protoReqHandlerMap: make(protoReqHandlerMap),
		Conn:               nil,
	}
	wsConnChan := make(chan *ws.Conn)
	ws.Connect(address, func(event *ws.Event, conn *ws.Conn) {
		debug("Client:", event)
		switch event.Type {
		case ws.Connected:
			wsConnChan <- conn
		case ws.BinaryMessage:
			client.Conn.readAndHandleWireWrapperReader(event)
		case ws.Disconnected:
			client.Log("TODO: reconnect logic (Disconnected)")
			if client.OnDisconnectHack != nil {
				client.OnDisconnectHack()
			}
		case ws.NetError:
			client.Log("NetError")
		default:
			panic("TODO Handle event: " + event.String())
		}
	})
	client.Conn = newConn(<-wsConnChan, client.jsonReqHandlerMap, client.protoReqHandlerMap)
	return
}
