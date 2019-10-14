package server

import (
	"encoding/binary"
	"fmt"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// XRequest device
type XRequest struct {
	uuid   string
	port   uint16
	conn   *websocket.Conn
	idx    uint16
	tag    uint16
	inUsed bool
	dev    *XDevice
}

func (r *XRequest) onData(data []byte) error {
	if r.conn != nil {
		return r.conn.WriteMessage(websocket.BinaryMessage, data)
	}

	return fmt.Errorf("XRequest no conn")
}

func (r *XRequest) close() {
	if r.conn != nil {
		r.conn.Close()
	}

	r.conn = nil
}

func (r *XRequest) free() {
	if r.conn != nil {
		r.conn.Close()
	}

	r.inUsed = false
	r.conn = nil
	r.tag++
	r.uuid = ""
	r.dev = nil

	log.Printf("xrequest free, idx:%d, tag:%d", r.idx, r.tag)
}

func (r *XRequest) use(uuid string, port uint16, conn *websocket.Conn, dev *XDevice) {
	r.dev = dev
	r.inUsed = true
	r.uuid = uuid
	r.tag++
	r.conn = conn
	r.port = port
}

func (r *XRequest) loopMsg() {
	c := r.conn
	if c == nil {
		log.Println("xrequest loopmsg failed, nil conn")
		return
	}

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("xrequest websocket read:", err)
			break
		}

		r.xClientData(message)
	}

	r.xClientClosed()
	r.free()
}

// xClientData x means exchange, send to server
func (r *XRequest) xClientData(message []byte) {
	if r.dev == nil {
		log.Println("failed to send client data, xdev is nil")
		return
	}

	dev := r.dev
	new := make([]byte, 5+len(message))
	new[0] = cmdReqData
	binary.LittleEndian.PutUint16(new[1:], r.idx)
	binary.LittleEndian.PutUint16(new[3:], r.tag)
	copy(new[5:], message)

	dev.sendMsg(new)
}

func (r *XRequest) xClientClosed() {
	if r.dev == nil {
		log.Println("failed to send client close, xdev is nil")
		return
	}

	dev := r.dev
	new := make([]byte, 5)
	new[0] = cmdReqClientClosed
	binary.LittleEndian.PutUint16(new[1:], r.idx)
	binary.LittleEndian.PutUint16(new[3:], r.tag)

	dev.sendMsg(new)
}

func (r *XRequest) xClientCreate() {
	if r.dev == nil {
		log.Println("failed to send client create, xdev is nil")
		return
	}

	dev := r.dev
	new := make([]byte, 7)
	new[0] = cmdReqCreated
	binary.LittleEndian.PutUint16(new[1:], r.idx)
	binary.LittleEndian.PutUint16(new[3:], r.tag)
	binary.LittleEndian.PutUint16(new[5:], r.port)

	dev.sendMsg(new)
}
