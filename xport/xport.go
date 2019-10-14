package server

import (
	"lproxy/server"
	"lproxy/xport/lws"
	"strconv"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var (
	upgrader    = websocket.Upgrader{} // use default options
	lwsupgrader = lws.Upgrader{}

	devices = make(map[string]*XDevice)
)

const (
	cmdNone              = 0
	cmdReqData           = 1
	cmdReqCreated        = 2
	cmdReqClientClosed   = 3
	cmdReqClientFinished = 4
	cmdReqServerFinished = 5
	cmdReqServerClosed   = 6
	cmdReqClientQuota    = 7
	cmdPing              = 8
	cmdPong              = 9
)

func xportServeLWS(ctx *server.RequestContext) {
	query := ctx.Query
	uuid := ctx.UUID
	if uuid == "" {
		ctx.Log.Panicln("invalid uuid")
	}

	capstr := query.Get("cap")
	cap, err := strconv.Atoi(capstr)
	if err != nil {
		ctx.Log.Panicln("convert cap error:", err)
	}

	c, err := lwsupgrader.Upgrade(ctx.W, ctx.R)
	if err != nil {
		ctx.Log.Panicln("upgrade:", err)
		return
	}

	peerAddr := c.RemoteAddr()
	ctx.Log.Println("accept lws from:", peerAddr)
	defer c.Close()

	// wait old xdevice to exit
	old, ok := devices[uuid]
	if ok {
		old.close()
		old.wg.Wait()
		ctx.Log.Println("wait old xdevice exit ok:", uuid)
	}

	new := newXDevice(uuid, c, cap)
	_, ok = devices[uuid]
	if ok {
		ctx.Log.Panicln("try to add device conflict")
	}

	devices[uuid] = new
	new.wg.Add(1)
	defer func() {
		delete(devices, uuid)
		new.wg.Done()
	}()

	new.loopMsg()
	ctx.Log.Println("serv lws end:", peerAddr)
}

func xportServeWebsocket(ctx *server.RequestContext) {
	devUUID := ctx.Query.Get("uuid")
	if devUUID == "" {
		log.Panicln("no dev uuid provided")
	}

	targetPortStr := ctx.Query.Get("port")
	if targetPortStr == "" {
		log.Panicln("no port provided")
	}

	targetPort, err := strconv.Atoi(targetPortStr)
	if err != nil {
		log.Panicln("convert port failed:", err)
	}

	xdev, ok := devices[devUUID]
	if !ok {
		log.Panicln("no dev found for uuid:", devUUID)
	}

	c, err := upgrader.Upgrade(ctx.W, ctx.R, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	log.Println("accept websocket from:", c.RemoteAddr())
	defer c.Close()

	xreq := xdev.mountRequest(devUUID, uint16(targetPort), c)
	if xreq != nil {
		xreq.loopMsg()
	} else {
		log.Panicln("failed to mount request into xdev")
	}
}

func init() {
	server.RegisterGetHandleNoUUID("/xportws", xportServeWebsocket)
	server.RegisterGetHandle("/xportlws", xportServeLWS)
}
