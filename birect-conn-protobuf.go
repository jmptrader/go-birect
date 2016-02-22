package birect

import (
	"github.com/golang/protobuf/proto"
	"github.com/marcuswestin/go-birect/internal/wire"
	"github.com/marcuswestin/go-errs"
)

// Proto is an alias for proto.Message
type Proto proto.Message

// ProtoReqHandler functions get called on every proto request
type ProtoReqHandler func(req *ProtoReq) (resValue Proto, err error)

// SendProtoReq sends a request for the ProtoReqHandler with the given `name`, along with the
// given paramsObj. When the server responds, SendProtoReq will parse the response into resValPtr.
func (c *Conn) SendProtoReq(name string, paramsObj Proto, resValPtr Proto) (err error) {
	data, err := proto.Marshal(paramsObj)
	if err != nil {
		return
	}
	reqID := c.nextReqID()
	wireReq := &wire.Request{Type: wire.DataType_Proto, Name: name, ReqId: uint32(reqID), Data: data}
	return c.sendRequestAndWaitForResponse(reqID, wireReq, resValPtr)
}

// ProtoReq wraps a request sent via SendProtoReq. Use ParseParams to access the proto values.
type ProtoReq struct {
	*Conn
	data []byte
}

// ParseParams parses the ProtoReq values into the given valuePtr.
// valuePtr should be a pointer to a struct that implements Proto.message.
func (p *ProtoReq) ParseParams(valuePtr Proto) {
	err := proto.Unmarshal(p.data, valuePtr)
	if err != nil {
		panic(errs.Wrap(err, nil, "Unable to parse params"))
	}
}

// Internal
///////////

type protoReqHandlerMap map[string]ProtoReqHandler

func (m protoReqHandlerMap) HandleProtoReq(reqName string, handler ProtoReqHandler) {
	m[reqName] = handler
}

func (c *Conn) handleProtoWireReq(wireReq *wire.Request) {
	// Find handler
	handler, exists := c.protoReqHandlerMap[wireReq.Name]
	if !exists {
		c.sendErrorResponse(wireReq, errs.New(nil, "Missing request handler"))
		return
	}
	// Execute handler
	resVal, err := handler(&ProtoReq{c, wireReq.Data})
	if err != nil {
		c.sendErrorResponse(wireReq, err)
		return
	}
	// Send response
	c.sendResponse(wireReq, &protoRes{resVal})
}

type protoRes struct {
	resValPtr Proto
}

func (j *protoRes) encode() ([]byte, error) {
	return proto.Marshal(j.resValPtr)
}
func (j *protoRes) dataType() wire.DataType {
	return wire.DataType_Proto
}
