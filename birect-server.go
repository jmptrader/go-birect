package birect

import (
	"net"
	"net/http"
	"sync"

	"github.com/marcuswestin/go-ws"
)

// Handler is used register request handlers (for requests sent from clients),
// and to accept incoming connections from birect clients.
type Handler struct {
	jsonReqHandlerMap
	protoReqHandlerMap
	connByWSConnMutex *sync.Mutex
	connByWSConn      map[*ws.Conn]*Conn
	ConnectHandler    func(*Conn)
	DisconnectHandler func(*Conn)
}

// UpgradeRequests will upgrade all incoming HTTP requests that match `pattern`
// to birect connections. Instead of using server.ListenAndServe(), you should
// call http.ListenAndServe()
func UpgradeRequests(pattern string) *Handler {
	handler := newHandler()
	ws.UpgradeRequests(pattern, getEventHandler(handler))
	return handler
}

// NewServer returns a new server that you are expected to
func NewServer() *Server {
	return &Server{newHandler()}
}

// Server allows you to create a standalone birect upgrade server.
// Call ListenAndServe to start upgrading all incoming http requests.
type Server struct {
	*Handler
}

func newHandler() *Handler {
	return &Handler{
		make(jsonReqHandlerMap),
		make(protoReqHandlerMap),
		&sync.Mutex{},
		make(map[*ws.Conn]*Conn, 10000),
		func(*Conn) {},
		func(*Conn) {},
	}
}

// ListenAndServe will start listening to the given address and upgrading
// any incoming http requests to websocket and birect connections.
func (s *Handler) ListenAndServe(address string) (errChan chan error) {
	errChan = make(chan error)
	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.UpgradeHandlerFunc(getEventHandler(s)))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		go func() {
			errChan <- err
		}()
		return
	}
	go func() {
		errChan <- http.Serve(listener, mux)
	}()
	return errChan
}

func getEventHandler(server *Handler) ws.EventHandler {
	return func(event *ws.Event, wsConn *ws.Conn) {
		switch event.Type {
		case ws.Connected:
			server.registerConn(wsConn)
		case ws.BinaryMessage:
			if conn := server.getConn(wsConn); conn != nil {
				conn.readAndHandleWireWrapperReader(event)
			}
		case ws.NetError:
			server.getConn(wsConn).Log("Net error")
		case ws.Disconnected:
			server.getConn(wsConn).Log("Disconnected")
			server.deregisterConn(wsConn)
		default:
			panic("birect.Handler unknown event: " + event.String())
		}
	}
}

// ConnCount returns the number of current connections
func (s *Handler) ConnCount() int {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	return len(s.connByWSConn)
}

// Conns returns all the current connections
func (s *Handler) Conns() (conns []*Conn) {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	conns = make([]*Conn, 0, len(s.connByWSConn))
	for _, conn := range s.connByWSConn {
		conns = append(conns, conn)
	}
	return
}

// Internal
///////////

func (s *Handler) registerConn(wsConn *ws.Conn) {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	conn := newConn(wsConn, s.jsonReqHandlerMap, s.protoReqHandlerMap)
	s.connByWSConn[wsConn] = conn
	if s.ConnectHandler != nil {
		defer s.ConnectHandler(conn)
	}
}
func (s *Handler) deregisterConn(wsConn *ws.Conn) {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	conn := s.connByWSConn[wsConn]
	delete(s.connByWSConn, wsConn)
	if s.DisconnectHandler != nil {
		defer s.DisconnectHandler(conn)
	}
}
func (s *Handler) getConn(wsConn *ws.Conn) *Conn {
	s.connByWSConnMutex.Lock()
	defer s.connByWSConnMutex.Unlock()
	return s.connByWSConn[wsConn]
}
