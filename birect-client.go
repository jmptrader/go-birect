package birect

import "github.com/marcuswestin/go-ws"

// Client is used register request handlers (for requests sent from the server),
// and to send requests to the server.
type Client struct {
	jsonReqHandlerMap
	protoReqHandlerMap
	*Conn
}

// Connect connects to a birect server at url
func Connect(url string) (client *Client, err error) {
	client = &Client{make(jsonReqHandlerMap), make(protoReqHandlerMap), nil}
	wsConnChan := make(chan *ws.Conn)
	ws.Connect(url, func(event *ws.Event, conn *ws.Conn) {
		debug("Client:", event)
		switch event.Type {
		case ws.Connected:
			wsConnChan <- conn
		case ws.BinaryMessage:
			client.Conn.readAndHandleWireWrapperReader(event)
		default:
			panic("TODO Handle event: " + event.String())
		}
	})
	client.Conn = newConn(<-wsConnChan, client.jsonReqHandlerMap, client.protoReqHandlerMap)
	return
}
