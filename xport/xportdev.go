package server

import (
	"encoding/binary"
	"lproxy/xport/lws"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// XDevice device
type XDevice struct {
	uuid     string
	conn     *lws.Conn
	requests []*XRequest
	wg       sync.WaitGroup
}

func newXDevice(uuid string, conn *lws.Conn, cap int) *XDevice {
	requests := make([]*XRequest, cap)
	for i := 0; i < cap; i++ {
		requests[i] = &XRequest{idx: uint16(i)}
	}

	return &XDevice{
		uuid:     uuid,
		conn:     conn,
		requests: requests,
	}
}

func (d *XDevice) close() {
	if d.conn != nil {
		d.conn.Close()
	}

	d.conn = nil
}

func (d *XDevice) free() {
	for _, r := range d.requests {
		r.close()
	}

	log.Printf("XDevice free, uuid:%s", d.uuid)
}

func (d *XDevice) sendMsg(msg []byte) {
	if d.conn != nil {
		err := d.conn.WriteMessage(msg)
		if err != nil {
			log.Println("XDevice sendMsg failed:", err)
		}
	}
}

func (d *XDevice) loopMsg() {
	c := d.conn
	for {
		message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// log.Println("recv msg length: ", len(message))
		cmd := message[0]
		if cmd == cmdPing {
			// log.Println("recv ping from peer, send pong")
			message[0] = cmdPong
			c.WriteMessage(message)
		} else if cmd == cmdPong {
			// nothing to do
		} else {
			d.handleRequestMsg(message)
		}
	}

	d.free()
}

func (d *XDevice) handleRequestMsg(message []byte) {
	if len(message) < 5 {
		log.Errorln("request message len should >= 5")
		return
	}

	cmd := message[0]
	requestIdx := binary.LittleEndian.Uint16(message[1:])
	requestTag := binary.LittleEndian.Uint16(message[3:])

	req := d.getRequest(requestIdx, requestTag)
	if req == nil {
		log.Printf("no request found for idx:%d tag:%d", requestIdx, requestTag)
		return
	}

	switch cmd {
	case cmdReqServerFinished:
		fallthrough
	case cmdReqServerClosed:
		req.close()
	case cmdReqData:
		err := req.onData(message[5:])
		if err != nil {
			log.Println("req.onData failed:", err)
			req.close()
		}

	default:
		log.Println("handleRequestMsg, unknown cmd:", cmd)
	}
}

func (d *XDevice) getRequest(requestIdx uint16, requestTag uint16) *XRequest {
	if int(requestIdx) >= len(d.requests) {
		return nil
	}

	req := d.requests[requestIdx]
	if req == nil {
		return nil
	}

	if req.tag != requestTag {
		return nil
	}

	if !req.inUsed {
		return nil
	}

	return req
}

func (d *XDevice) mountRequest(uuid string, targetPort uint16, conn *websocket.Conn) *XRequest {
	var req *XRequest
	for _, r := range d.requests {
		if !r.inUsed {
			req = r
			break
		}
	}

	if req == nil {
		return nil
	}

	req.use(uuid, targetPort, conn, d)
	req.xClientCreate()

	return req
}
