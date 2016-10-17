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
	ConnectHandler    func(*Conn)
	DisconnectHandler func(*Conn)
}

// UpgradeRequests will upgrade all incoming HTTP requests that match `pattern`
// to birect connections.
func UpgradeRequests(pattern string) (server *Server) {
	server = &Server{
		make(jsonReqHandlerMap),
		make(protoReqHandlerMap),
		&sync.Mutex{},
		make(map[*ws.Conn]*Conn, 10000),
		func(*Conn) {},
		func(*Conn) {},
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
	conn := newConn(wsConn, s.jsonReqHandlerMap, s.protoReqHandlerMap)
	s.connByWSConn[wsConn] = conn
	if s.ConnectHandler != nil {
		defer s.ConnectHandler(conn)
	}
}
func (s *Server) deregisterConn(wsConn *ws.Conn) {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	conn := s.connByWSConn[wsConn]
	delete(s.connByWSConn, wsConn)
	if s.DisconnectHandler != nil {
		defer s.DisconnectHandler(conn)
	}
}
func (s *Server) getConn(wsConn *ws.Conn) *Conn {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	return s.connByWSConn[wsConn]
}
