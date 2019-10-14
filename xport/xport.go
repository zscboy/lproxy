package server

import (
	"lproxy/server"
	"lproxy/servercfg"
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
		ctx.Log.Println("invalid uuid")
		return
	}

	capstr := query.Get("cap")
	cap, err := strconv.Atoi(capstr)
	if err != nil {
		ctx.Log.Println("convert cap error:", err)
		return
	}

	c, err := lwsupgrader.Upgrade(ctx.W, ctx.R)
	if err != nil {
		ctx.Log.Println("upgrade:", err)
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
		ctx.Log.Println("try to add device conflict")
		return
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
		log.Println("no dev uuid provided")
		return
	}

	targetPortStr := ctx.Query.Get("port")
	if targetPortStr == "" {
		log.Println("no port provided")
		return
	}

	targetPort, err := strconv.Atoi(targetPortStr)
	if err != nil {
		log.Println("convert port failed:", err)
		return
	}

	xdev, ok := devices[devUUID]
	if !ok {
		log.Println("no dev found for uuid:", devUUID)
		return
	}

	c, err := upgrader.Upgrade(ctx.W, ctx.R, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	log.Println("accept websocket from:", c.RemoteAddr())
	defer c.Close()

	xreq := xdev.mountRequest(devUUID, uint16(targetPort), c)
	if xreq != nil {
		xreq.loopMsg()
	} else {
		log.Println("failed to mount request into xdev")
		return
	}
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		server.RegisterGetHandleNoUUID(servercfg.XPortWebsocketPath, xportServeWebsocket)
		server.RegisterGetHandle(servercfg.XPortLWSPath, xportServeLWS)
	})
}
