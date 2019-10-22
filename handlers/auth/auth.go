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

	Arch string `json:"arch"`
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
	response.Token = token

	handleUpgrade(req.Arch, req.Version, response)
	handleDomains("", response)

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

	if len(bytesArray) >= 4*1024 {
		acceptContentEncodeStr := ctx.R.Header.Get("Accept-Encoding")

		if strings.Contains(acceptContentEncodeStr, "gzip") {
			log.Println("client support gzip")
			gzipSupport = true
		}
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

func handleUpgrade(arch string, currentVerStr string, response *Response) {
	fm, ok := servercfg.FirmwareMap[arch]
	if !ok {
		// no upgrade config
		return
	}

	needUpgrade := true
	if semverLE(fm.NewVersion, currentVerStr) {
		needUpgrade = false
	} else {
		log.Printf("authHandle, client ver:%s old than current:%s, upgrade", currentVerStr,
			fm.NewVersionStr)
	}

	response.NeedUpgrade = needUpgrade
	response.UpgradeURL = fm.UpgradeURL
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		server.RegisterPostHandleNoUUID(servercfg.AuthPath, authHandle)
	})
}
