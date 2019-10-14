package servercfg

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

var (
	tuncfg *TunCfg
)

// GetTunCfg get tunnel cfg
func GetTunCfg() *TunCfg {
	return tuncfg
}

// TunCfg tunnel cfg
type TunCfg struct {
	TunnelNumber    int      `json:"tunnel_number"`
	WebsocketURL    string   `json:"websocket_url"`
	DNSTunURL       string   `json:"dns_tun_url"`
	LocalServer     string   `json:"local_server"`
	TunnelReqCap    int      `json:"tunnel_req_cap"`
	RelayDomain     string   `json:"relay_domain"`
	RelayPort       int      `json:"relay_port"`
	LocalTCPPort    int      `json:"local_tcp_port"`
	DNSTunnelNumber int      `json:"dns_tunnel_number"`
	LocalDNSServer  string   `json:"local_dns_server"`
	XPortURL        string   `json:"xport_url"`
	Domains         []string `json:"domain_array"`
}

func loadTunCfgFromFile(filepath string) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	tcfg := &TunCfg{}
	err = json.Unmarshal(content, tcfg)
	if err != nil {
		log.Fatal(err)
	}
	tuncfg = tcfg
}
