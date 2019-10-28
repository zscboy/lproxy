package auth

import (
	"encoding/json"
	"lproxy/server"
	"lproxy/servercfg"

	"github.com/blang/semver"
	log "github.com/sirupsen/logrus"
)

// CfgMonitorRequest cfg monitor request
type CfgMonitorRequest struct {
	Version string `json:"current_version"`
	Arch    string `json:"arch"`

	DomainsVer string `json:"domains_ver"`
}

func cfgMonitorHandle(ctx *server.RequestContext) {
	log.Println("cfgMonitorHandle, uuid:", ctx.UUID)

	body := ctx.Body
	req := &CfgMonitorRequest{}
	err := json.Unmarshal(body, req)
	if err != nil {
		ctx.Log.Println("authHandle, unmarshal body failed:", err)
		return
	}

	response := &Response{}
	response.Error = 0
	response.Restrart = false
	response.NeedUpgrade = false

	response.Token = ctx.Query.Get("tok")

	handleUpgrade(req.Arch, req.Version, response)
	handleDomains(req.DomainsVer, response)

	b, err := json.Marshal(response)
	if err != nil {
		ctx.Log.Println("authHandle, Marshal response failed:", err)
		return
	}

	// zip
	writeHTTPBodyWithGzip(ctx, b)
}

func semverLE(v1v semver.Version, v2 string) bool {
	v2v, e := semver.Make(v2)
	if e != nil {
		log.Println("semver.NewVersion failed:", e)
		return false
	}

	return v1v.Compare(v2v) <= 0
}

func handleDomains(domainVer string, response *Response) {
	needDomains := true
	if domainVer != "" {
		currentDomainsVer := domainVer
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

func init() {
	server.InvokeAfterCfgLoaded(func() {
		server.RegisterPostHandle(servercfg.CfgMonitorPath, cfgMonitorHandle)
	})
}
