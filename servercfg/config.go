package servercfg

import (
	"encoding/json"
	"os"

	"github.com/blang/semver"
	log "github.com/sirupsen/logrus"

	"github.com/DisposaBoy/JsonConfigReader"
)

// make a copy of this file, rename to settings.go
// then set the correct value for these follow variables
var (
	ServerPort = 8000
	// LogFile              = ""
	Daemon      = "yes"
	RedisServer = ":6379"
	ServerID    = ""

	ForTestOnly = false

	AsHTTPS     = true
	PfxLocation = "/home/abc/identity.pfx"
	PfxPassword = "123456"

	XPortLWSPath       = "/xportlws"
	XPortWebsocketPath = "/xportws"
	AuthPath           = "/auth"
	CfgMonitorPath     = "/cfgmonitor"

	BandwidthKbs = 0

	TokenKey = "@yymmxxkk#$yzilm"

	FirmwareMap = make(map[string]*FirmwareVersion)
)

var (
	loadedCfgFilePath = ""
)

// FirmwareVersion firemware config
type FirmwareVersion struct {
	Arch          string `json:"arch"`
	UpgradeURL    string `json:"upgrade_url"`
	NewVersionStr string `json:"new_version"`

	NewVersion semver.Version
}

// ReLoadConfigFile 重新加载配置
func ReLoadConfigFile() bool {
	log.Println("ReLoadConfigFile-----------From File--------:", loadedCfgFilePath)
	if !ParseConfigFile(loadedCfgFilePath) {
		log.Println("ReLoadConfigFile-------------------FAILED")
		return false
	}

	log.Println("ReLoadConfigFile-------------------OK")
	return true
}

// ParseConfigFile 解析配置
func ParseConfigFile(filepath string) bool {
	type Params struct {
		ServerPort  int    `json:"port"`
		Daemon      string `json:"daemon"`
		RedisServer string `json:"redis_server"`
		ServreID    string `json:"guid"`

		DomiansFile string `json:"domainsfile"`
		TunCfgFile  string `json:"tuncfgfile"`

		PfxLocation        string `json:"pfx_location"`
		PfxPassword        string `json:"pfx_password"`
		XPortLWSPath       string `json:"xport_lwspath"`
		XPortWebsocketPath string `json:"xport_wspath"`

		AsHTTPS  bool   `json:"as_https"`
		AuthPath string `json:"auth_path"`

		CfgMonitorPath string `json:"cfg_monitor_path"`

		TokenKey string `json:"token_key"`

		FirmwareArray []*FirmwareVersion `json:"firmwares"`

		BandwidthKbs int `json:"bandwidth_kbs"`
	}

	loadedCfgFilePath = filepath

	var params = &Params{}

	f, err := os.Open(filepath)
	if err != nil {
		log.Println("failed to open config file:", filepath)
		return false
	}

	// wrap our reader before passing it to the json decoder
	r := JsonConfigReader.New(f)
	err = json.NewDecoder(r).Decode(params)

	if err != nil {
		log.Println("json un-marshal error:", err)
		return false
	}

	log.Println("-------------------Configure params are:-------------------")
	log.Printf("%+v\n", params)

	// if params.LogFile != "" {
	// 	LogFile = params.LogFile
	// }

	if params.Daemon != "" {
		Daemon = params.Daemon
	}

	if params.ServerPort != 0 {
		ServerPort = params.ServerPort
	}

	if params.RedisServer != "" {
		RedisServer = params.RedisServer
	}

	if params.ServreID != "" {
		ServerID = params.ServreID
	}

	if ServerID == "" {
		log.Println("Server id 'guid' must not be empty!")
		return false
	}

	if RedisServer == "" {
		log.Println("redis server id  must not be empty!")
		return false
	}

	if params.DomiansFile != "" {
		loadDomainsFromFile(params.DomiansFile)
	}

	if params.TunCfgFile != "" {
		loadTunCfgFromFile(params.TunCfgFile)
	}

	if params.PfxLocation != "" {
		PfxLocation = params.PfxLocation
	}

	if params.PfxPassword != "" {
		PfxPassword = params.PfxPassword
	}

	if params.XPortLWSPath != "" {
		XPortLWSPath = params.XPortLWSPath
	}

	if params.XPortWebsocketPath != "" {
		XPortWebsocketPath = params.XPortWebsocketPath
	}

	if params.AuthPath != "" {
		AuthPath = params.AuthPath
	}

	if params.CfgMonitorPath != "" {
		CfgMonitorPath = params.CfgMonitorPath
	}

	BandwidthKbs = params.BandwidthKbs
	AsHTTPS = params.AsHTTPS

	if len(params.FirmwareArray) > 0 {
		for _, f := range params.FirmwareArray {
			var e error
			f.NewVersion, e = semver.Make(f.NewVersionStr)
			if e != nil {
				log.Fatalln("Config parse, failed to convert new version:", e)
			}

			FirmwareMap[f.Arch] = f
		}

		log.Printf("config FirmwareMap:%+v", FirmwareMap)
	}

	return true
}
