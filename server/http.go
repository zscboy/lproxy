package server

import (
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"lproxy/servercfg"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"io/ioutil"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"golang.org/x/crypto/pkcs12"
)

var (
	// 根router，只有http server看到
	rootRouter = httprouter.New()

	rootPath = ""
)

// GetVersion server version string
func GetVersion() string {
	return "0.1.0"
}

// GetVersionCode server version code
func GetVersionCode() int {
	return 1
}

// CreateHTTPServer 启动服务器
func CreateHTTPServer() {
	log.Printf("CreateHTTPServer")

	rootRouter.Handle("GET", rootPath+"/version", echoVersion)
	go acceptHTTPRequest()
}

func echoVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte(fmt.Sprintf("version:%s", GetVersion())))
}

// acceptHTTPRequest 监听和接受HTTP
func acceptHTTPRequest() {
	var hh http.Handler
	if servercfg.ForTestOnly {
		// 支持客户端跨域访问
		c := cors.New(cors.Options{
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			AllowCredentials: true,
			AllowedHeaders:   []string{"*"},           // we need this line for cors to allow cross-origin
			ExposedHeaders:   []string{"Set-Session"}, // we need this line for cors to set Access-Control-Expose-Headers
		})
		hh = c.Handler(rootRouter)
	} else {
		// 对外服务器不应该允许跨域访问
		hh = rootRouter
	}

	portStr := fmt.Sprintf(":%d", servercfg.ServerPort)
	if servercfg.AsHTTPS {
		prxdata := loadPfx()
		blocks, err := pkcs12.ToPEM(prxdata, servercfg.PfxPassword)
		if err != nil {
			log.Fatalln("ToPEM failed:", err)
		}

		var pemData []byte
		for _, b := range blocks {
			pemData = append(pemData, pem.EncodeToMemory(b)...)
		}

		// then use PEM data for tls to construct tls certificate:
		cert, err := tls.X509KeyPair(pemData, pemData)
		if err != nil {
			log.Fatalln("X509KeyPair failed:", err)
		}

		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		s := &http.Server{
			Addr:           portStr,
			Handler:        hh,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 10,
			TLSConfig:      config,
			TLSNextProto:   make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		log.Printf("Https server listen at:%d\n", servercfg.ServerPort)

		err = s.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatalf("Http server ListenAndServe %d failed:%s\n", servercfg.ServerPort, err)
		}
	} else {
		s := &http.Server{
			Addr:           portStr,
			Handler:        hh,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 10,
		}

		log.Printf("Http server listen at:%d\n", servercfg.ServerPort)

		err := s.ListenAndServe()
		if err != nil {
			log.Fatalf("Http server ListenAndServe %d failed:%s\n", servercfg.ServerPort, err)
		}
	}
}

func loadPfx() []byte {
	data, err := ioutil.ReadFile(servercfg.PfxLocation)
	if err != nil {
		log.Fatalln("loadPfx failed:", err)
	}

	return data
}
