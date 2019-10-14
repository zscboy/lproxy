package lws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Upgrader lws upgrader
type Upgrader struct {
}

// tokenListContainsValue returns true if the 1#token header with the given
// name contains a token equal to value with ASCII case folding.
func tokenListContainsValue(header http.Header, name string, value string) bool {
	for _, s := range header[name] {
		if strings.Contains(s, value) {
			return true
		}
	}

	return false
}

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func computeAcceptKey(challengeKey string) string {
	h := sha1.New()
	h.Write([]byte(challengeKey))
	h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Upgrade upgrade http request to lws conn
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		return nil, fmt.Errorf("'upgrade' token not found in 'Connection' header")
	}

	if !tokenListContainsValue(r.Header, "Upgrade", "websocket") {
		return nil, fmt.Errorf("'websocket' token not found in 'Upgrade' header")
	}

	if r.Method != "GET" {
		return nil, fmt.Errorf("request method is not GET")
	}

	if !tokenListContainsValue(r.Header, "Sec-Websocket-Version", "13") {
		return nil, fmt.Errorf("websocket: unsupported version: 13 not found in 'Sec-Websocket-Version' header")
	}

	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("websocket: response does not implement http.Hijacker")
	}

	var brw *bufio.ReadWriter
	netConn, brw, err := h.Hijack()
	if err != nil {
		return nil, err
	}

	if brw.Reader.Buffered() > 0 {
		netConn.Close()
		return nil, fmt.Errorf("websocket: client sent data before handshake is complete")
	}

	c := newConn(netConn)

	challengeKey := r.Header.Get("Sec-Websocket-Key")
	if challengeKey == "" {
		netConn.Close()
		return nil, fmt.Errorf("websocket: not a websocket handshake: 'Sec-WebSocket-Key' header is missing or blank")
	}

	var p []byte = nil
	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, computeAcceptKey(challengeKey)...)
	p = append(p, "\r\n"...)
	p = append(p, "\r\n"...)

	// Clear deadlines set by HTTP server.
	netConn.SetDeadline(time.Time{})

	if _, err = netConn.Write(p); err != nil {
		netConn.Close()
		return nil, err
	}

	return c, nil
}

// Conn lws connection
type Conn struct {
	nc         net.Conn
	writeMutex sync.Mutex
}

func newConn(nc net.Conn) *Conn {
	return &Conn{nc: nc}
}

// Close close lws connection
func (c *Conn) Close() error {
	return c.nc.Close()
}

// ReadMessage read message from lws connection
func (c *Conn) ReadMessage() ([]byte, error) {
	// read 2 bytes header
	lenBuf := []byte{0, 0}
	_, err := io.ReadFull(c.nc, lenBuf)
	if err != nil {
		return nil, err
	}

	var len uint16 = uint16(lenBuf[1])
	len = (len << 8) | uint16(lenBuf[0])
	if len <= 2 {
		return nil, fmt.Errorf("lws invalid length")
	}

	len = len - 2
	content := make([]byte, len)
	_, err = io.ReadFull(c.nc, content)
	if err != nil {
		return nil, err
	}

	// read remain content
	return content, nil
}

// WriteMessage write message to lws connection
func (c *Conn) WriteMessage(content []byte) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// write 2 bytes header
	len := uint16(len(content) + 2)
	lenBuf := []byte{0, 0}
	lenBuf[1] = byte(len >> 8)
	lenBuf[0] = byte(len)
	err := writeAll(lenBuf, c.nc)
	if err != nil {
		return err
	}

	// write remain content
	return writeAll(content, c.nc)
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.nc.RemoteAddr()
}

func writeAll(buf []byte, nc net.Conn) error {
	wrote := 0
	l := len(buf)
	for {
		n, err := nc.Write(buf[wrote:])
		if err != nil {
			return err
		}

		wrote = wrote + n
		if wrote == l {
			break
		}
	}

	return nil
}
