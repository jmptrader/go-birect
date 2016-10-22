package birect

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/marcuswestin/go-birect/internal/wire"
	"github.com/marcuswestin/go-errs"
	"github.com/marcuswestin/go-ws"
)

// Debug will generate verbose output when set to true.
var Debug = false

func debug(args ...interface{}) {
	if Debug {
		log.Println(args...)
	}
}

// Conn represents a persistent bi-directional connection between
// a birect client and a birect server.
type Conn struct {
	wsConn    *ws.Conn
	lastReqID reqID
	resChans  map[reqID]resChan
	jsonReqHandlerMap
	protoReqHandlerMap

	// Use info to associate information with a connection, e.g
	// authenticated user information
	Info interface{}
}

// Log logs the given arguments, along with contextual information about the Conn.
func (c *Conn) Log(args ...interface{}) {
	debug(args...)
}

// Internal
///////////

type reqID uint32
type resChan chan *wire.Response

func newConn(wsConn *ws.Conn, jsonHandlers jsonReqHandlerMap, protoHandlers protoReqHandlerMap) *Conn {
	return &Conn{wsConn, 0, make(map[reqID]resChan, 1), jsonHandlers, protoHandlers, nil}
}

type request interface {
	// Request sending side
	encode() ([]byte, error)
	// Request handling side
	ParseParams(valPrt interface{})
}
type response interface {
	// Responding side
	dataType() wire.DataType
	encode() ([]byte, error)
	// Response receiving side
}

// Internal - Outgoing wrappers
///////////////////////////////

func (c *Conn) sendRequestAndWaitForResponse(reqID reqID, wireReq *wire.Request, resValPtr interface{}) (err error) {
	c.resChans[reqID] = make(resChan)
	defer delete(c.resChans, reqID)

	c.Log("REQ", wireReq.Name, "ReqID:", reqID, "len:", len(wireReq.Data))
	err = c.sendWrapper(&wire.Wrapper{
		Content: &wire.Wrapper_Request{Request: wireReq},
	})
	if err != nil {
		return
	}

	wireRes := <-c.resChans[reqID]
	c.Log("RCV", wireReq.Name, "ReqID:", reqID, "DataType:", wireRes.Type, "len(Data):", len(wireRes.Data))

	if wireRes.IsError {
		return errors.New(string(wireRes.Data))
	}

	if wireRes.Data == nil {
		return nil
	}

	switch wireRes.Type {
	case wire.DataType_JSON:
		return json.Unmarshal(wireRes.Data, resValPtr)
	case wire.DataType_Proto:
		return proto.Unmarshal(wireRes.Data, resValPtr.(proto.Message))
	default:
		return errors.New("Bad response wire type: " + wireRes.Type.String())
	}
}
func (c *Conn) sendResponse(wireReq *wire.Request, response response) {
	wireRes := &wire.Response{ReqId: wireReq.ReqId}
	data, err := response.encode()
	if err != nil {
		panic(errs.Wrap(err, nil, "Unable to encode response"))
	}
	wireRes.Type = response.dataType()
	wireRes.Data = data
	err = c.sendWrapper(&wire.Wrapper{
		Content: &wire.Wrapper_Response{Response: wireRes},
	})
	if err != nil {
		panic(errs.Wrap(err, nil, "Unable to send response"))
	}
}
func (c *Conn) sendErrorResponse(wireReq *wire.Request, err error) {
	var publicMessage string
	if errsErr, ok := err.(errs.Err); ok {
		publicMessage = errsErr.PublicMsg()
	}
	if publicMessage == "" {
		publicMessage = DefaultPublicErrorMessage
	}
	wireRes := &wire.Response{
		ReqId:   wireReq.ReqId,
		IsError: true,
		Type:    wire.DataType_Text,
		Data:    []byte(publicMessage),
	}
	c.Log("Req ERROR", wireReq.ReqId, err)
	c.sendWrapper(&wire.Wrapper{
		Content: &wire.Wrapper_Response{Response: wireRes},
	})
}
func (c *Conn) nextReqID() reqID {
	rawReqID := atomic.AddUint32((*uint32)(&c.lastReqID), 1)
	return reqID(rawReqID)
}
func (c *Conn) sendWrapper(wrapper *wire.Wrapper) (err error) {
	wireData, err := proto.Marshal(wrapper)
	if err != nil {
		return
	}
	c.Log("SND Wrapper len:", len(wireData), wrapper)
	return c.wsConn.SendBinary(wireData)
}

// Internal - incoming wrappers
///////////////////////////////

func (c *Conn) readAndHandleWireWrapperReader(reader io.Reader) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(errs.Wrap(err, nil, "Unable to read wrapper"))
	}
	c.readAndHandleWireWrapper(data)
}
func (c *Conn) readAndHandleWireWrapper(data []byte) {
	if len(data) == 0 {
		panic(errs.New(nil, "Empty data"))
	}

	var wireWrapper wire.Wrapper
	err := proto.Unmarshal(data, &wireWrapper)
	if err != nil {
		panic(errs.Wrap(err, nil, "Unable to decode wire wrapper"))
	}

	c.Log("readAndHandleWireWrapper", wireWrapper.Content)
	switch content := wireWrapper.Content.(type) {
	case *wire.Wrapper_Message:
		c.handleMessage(content.Message)
	case *wire.Wrapper_Request:
		c.handleRequest(content.Request)
	case *wire.Wrapper_Response:
		c.handleResponse(content.Response)
	default:
		panic(errs.New(errs.Info{"Wrapper": wireWrapper}, "Unknown wire wrapper content type"))
	}
}
func (c *Conn) handleMessage(msg *wire.Message) {
	panic(errs.New(nil, "TODO: handleMessage"))
}
func (c *Conn) handleRequest(wireReq *wire.Request) {
	c.Log("HANDLE REQ", wireReq)
	switch wireReq.Type {
	case wire.DataType_JSON:
		c.handleJSONWireReq(wireReq)
	case wire.DataType_Proto:
		c.handleProtoWireReq(wireReq)
	default:
		c.sendErrorResponse(wireReq, errs.New(errs.Info{"Type": wireReq.Type}, "Bad wireReq.Type"))
	}
}
func (c *Conn) handleResponse(wireRes *wire.Response) {
	c.Log("HANDLE RES", wireRes)
	c.resChans[reqID(wireRes.ReqId)] <- wireRes
}
