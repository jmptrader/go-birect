package birect

import (
	"log"
	"sync"

	"github.com/marcuswestin/go-ws"
)

// Server is used register request handlers (for requests sent from clients),
// and to accept incoming connections from birect clients.
type Server struct {
	jsonReqHandlerMap
	protoReqHandlerMap
	connByWSConnMutex *sync.Mutex
	connByWSConn      map[*ws.Conn]*Conn
}

// UpgradeRequests will upgrade all incoming HTTP requests that match `pattern`
// to birect connections.
func UpgradeRequests(pattern string) (server *Server) {
	server = &Server{
		make(jsonReqHandlerMap),
		make(protoReqHandlerMap),
		&sync.Mutex{},
		make(map[*ws.Conn]*Conn, 10000),
	}
	ws.UpgradeRequests(pattern, func(event *ws.Event, wsConn *ws.Conn) {
		log.Println("Server:", event)
		switch event.Type {
		case ws.Connected:
			server.registerConn(wsConn)
		case ws.BinaryMessage:
			if conn := server.getConn(wsConn); conn != nil {
				conn.readAndHandleWireWrapperReader(event)
			}
		case ws.Disconnected:
			server.deregisterConn(wsConn)
		default:
			panic("birect.Server unknown event: " + event.String())
		}
	})
	return server
}

// Internal
///////////

func (s *Server) registerConn(wsConn *ws.Conn) {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	s.connByWSConn[wsConn] = newConn(wsConn, s.jsonReqHandlerMap, s.protoReqHandlerMap)
}
func (s *Server) deregisterConn(wsConn *ws.Conn) {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	delete(s.connByWSConn, wsConn)
}
func (s *Server) getConn(wsConn *ws.Conn) *Conn {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	return s.connByWSConn[wsConn]
}
