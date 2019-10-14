package servercfg

import (
	"encoding/json"
	"os"

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

	DbIP        = "localhost"
	DbPort      = 1433
	DbUser      = "abc"
	DbPassword  = "ab"
	DbName      = "gamedb"
	ForTestOnly = false
	AsHTTPS     = true
	PfxLocation = "/home/abc/identity.pfx"
	PfxPassword = "123456"

	XPortLWSPath       = "/xportlws"
	XPortWebsocketPath = "/xportws"
)

var (
	loadedCfgFilePath = ""
)

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
		DbIP        string `json:"dbIP"`
		DbPort      int    `json:"dbPort"`
		DbPassword  string `json:"dbPassword"`
		DbUser      string `json:"dbUser"`
		DbName      string `json:"dbName"`

		DomiansFile string `json:"domainsfile"`
		TunCfgFile  string `json:"tuncfgfile"`
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

	if params.DbIP != "" {
		DbIP = params.DbIP
	}

	if params.DbUser != "" {
		DbUser = params.DbUser
	}

	if params.DbPassword != "" {
		DbPassword = params.DbPassword
	}

	if params.DbName != "" {
		DbName = params.DbName
	}

	if params.DbPort != 0 {
		DbPort = params.DbPort
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
		tuncfg.Domains = domains
	}

	return true
}
