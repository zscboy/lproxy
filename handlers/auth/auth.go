package auth

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"lproxy/server"
	"lproxy/servercfg"
	"strings"

	"github.com/blang/semver"
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

	IsForCfgMonitor bool   `json:"is_cfgmonitor"`
	DomainsVer      string `json:"domains_ver"`
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

	handleUpgrade(req, response)
	handleDomains(req, response)

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

func semverLE(v1v semver.Version, v2 string) bool {
	v2v, e := semver.Make(v2)
	if e != nil {
		log.Println("semver.NewVersion failed:", e)
		return false
	}

	return v1v.Compare(v2v) <= 0
}

func handleDomains(req *Request, response *Response) {
	needDomains := true
	if req.IsForCfgMonitor {
		currentDomainsVer := req.DomainsVer
		if semverLE(servercfg.DomainsCfgVer, currentDomainsVer) {
			needDomains = false
		} else {
			log.Printf("authHandle, client cfg monitor, domain ver:%s old than current:%s, update",
				currentDomainsVer, servercfg.DomainsCfgVerStr)
		}
	}

	response.TunCfg = servercfg.GetTunCfg()
	response.TunCfg.DomainsVer = servercfg.DomainsCfgVerStr

	if needDomains {
		response.TunCfg.Domains = servercfg.GetDomains()
	}
}

func handleUpgrade(req *Request, response *Response) {
	if servercfg.UpgradeURL == "" {
		// no upgrade config
		return
	}

	needUpgrade := true
	currentVer := req.Version
	if semverLE(servercfg.NewVersion, currentVer) {
		needUpgrade = false
	} else {
		log.Printf("authHandle, client ver:%s old than current:%s, upgrade", currentVer,
			servercfg.NewVersionStr)
	}

	response.NeedUpgrade = needUpgrade
	response.UpgradeURL = servercfg.UpgradeURL
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		server.RegisterPostHandleNoUUID(servercfg.AuthPath, authHandle)
	})
}
