package auth

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"lproxy/server"
	"lproxy/servercfg"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Response auth response
type Response struct {
	Error       int               `json:"error"`
	Token       string            `json:"token"`
	Restrart    bool              `json:"restart"`
	NeedUpgrade bool              `json:"need_upgrade"`
	UpgradeURL  string            `json:"upgrade_url,omitempty"`
	TunCfg      *servercfg.TunCfg `json:"tuncfg,omitempty"`
}

// Request auth request
type Request struct {
	UUID    string `json:"uuid"`
	Version string `json:"current_version"`
}

func authHandle(ctx *server.RequestContext) {
	body := ctx.Body
	req := &Request{}
	err := json.Unmarshal(body, req)
	if err != nil {
		ctx.Log.Println("authHandle, unmarshal body failed:", err)
		return
	}

	response := &Response{}
	response.Error = 0
	response.Restrart = false
	response.NeedUpgrade = false

	token := server.GenTK(req.UUID)
	response.TunCfg = servercfg.GetTunCfg()
	response.Token = token
	b, err := json.Marshal(response)
	if err != nil {
		ctx.Log.Println("authHandle, Marshal response failed:", err)
		return
	}

	// zip
	writeHTTPBodyWithGzip(ctx, b)
}

func writeHTTPBodyWithGzip(ctx *server.RequestContext, bytesArray []byte) {
	gzipSupport := false

	acceptContentEncodeStr := ctx.R.Header.Get("Accept-Encoding")

	if strings.Contains(acceptContentEncodeStr, "gzip") {
		log.Println("client support gzip")
		gzipSupport = true
	}

	if gzipSupport {
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		if _, err := g.Write(bytesArray); err != nil {
			log.Println("writeHTTPBodyWithGzip, write gzip err:", err)
			return
		}
		if err := g.Close(); err != nil {
			log.Println("writeHTTPBodyWithGzip, close gzip err:", err)
			return
		}

		ctx.W.Header().Set("Content-Encoding", "gzip")
		ctx.W.Header().Set("Content-Type", "application/octet-stream")

		bytesCompressed := buf.Bytes()
		log.Printf("COMPRESS, before:%d, after:%d\n", len(bytesArray), len(bytesCompressed))
		ctx.W.Write(bytesCompressed)
	} else {
		ctx.W.Write(bytesArray)
	}
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		server.RegisterPostHandleNoUUID(servercfg.AuthPath, authHandle)
	})
}
