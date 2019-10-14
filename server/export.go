package server

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// RequestContext simple http request context
type RequestContext struct {
	// Log 日志输出
	Log *log.Entry

	// UUID 当前请求的unique ID
	UUID string

	// Query cached query string
	// 余下的处理代码应该复用该query
	Query url.Values

	// Params httprouter params
	Params httprouter.Params

	// body bytes array, if exist
	Body []byte

	R *http.Request

	W http.ResponseWriter
}

// RequestHandle stupid handle
type RequestHandle func(*RequestContext)

// newReqContext 新建一个context
func newReqContext(r *http.Request, requiredUUID bool) *RequestContext {
	// parse UUID from token, if exist
	UUID := ""

	query := r.URL.Query()
	tk := query.Get("tok")
	if requiredUUID {
		var errCode int
		// try to parse token to get UUID
		UUID, errCode = parseTK(tk)
		if errCode != errTokenSuccess {
			return nil
		}
	}

	// construct context
	ctx := &RequestContext{}
	ctx.UUID = UUID
	ctx.Query = query
	// TODO: with or without IP address?
	ctx.Log = log.WithField("uuid", UUID)

	return ctx
}

// wrapGetHandleInternal 包装 get handle
func wrapGetHandleInternal(handle RequestHandle, requiredUUID bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := newReqContext(r, requiredUUID)
		if ctx == nil {
			return
		}

		ctx.Params = params

		ctx.R = r
		ctx.W = w
		handle(ctx)
	}
}

// wrapPostHandleInternal 包装 post handle
func wrapPostHandleInternal(handle RequestHandle, requiredUUID bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := newReqContext(r, requiredUUID)
		if ctx == nil {
			return
		}

		ctx.Params = params

		// read all body bytes
		// Read body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		ctx.Body = b
		ctx.R = r
		ctx.W = w

		handle(ctx)
	}
}

// RegisterGetHandle 注册http get handle
func RegisterGetHandle(subPath string, handle RequestHandle) {
	log.Info("RegisterGetHandle:", subPath)
	if subPath[0] != '/' {
		log.Panic("subPath must begin with '/', :", subPath)
	}

	path := rootPath + subPath
	h, _, _ := rootRouter.Lookup("GET", path)
	if h != nil {
		log.Panic("subPath with 'GET' has been register, subPath:", subPath)
	}

	rootRouter.GET(path, wrapGetHandleInternal(handle, true))
}

// RegisterPostHandle 注册http post handle
func RegisterPostHandle(subPath string, handle RequestHandle) {
	log.Info("RegisterPostHandle:", subPath)
	if subPath[0] != '/' {
		log.Panic("RegisterPostHandle subPath must begin with '/', :", subPath)
	}

	path := rootPath + subPath
	h, _, _ := rootRouter.Lookup("POST", path)
	if h != nil {
		log.Panic("RegisterPostHandle subPath with 'POST' has been register, subPath:", subPath)
	}

	rootRouter.POST(path, wrapPostHandleInternal(handle, true))
}

// RegisterPostHandleNoUUID 注册http post handle
func RegisterPostHandleNoUUID(subPath string, handle RequestHandle) {
	log.Info("RegisterPostHandleNoUUID:", subPath)
	if subPath[0] != '/' {
		log.Panic("RegisterPostHandleNoUUID subPath must begin with '/', :", subPath)
	}

	path := rootPath + subPath
	h, _, _ := rootRouter.Lookup("POST", path)
	if h != nil {
		log.Panic("RegisterPostHandleNoUUID subPath with 'POST' has been register, subPath:", subPath)
	}

	rootRouter.POST(path, wrapPostHandleInternal(handle, false))
}

// RegisterGetHandleNoUUID 注册http post handle
func RegisterGetHandleNoUUID(subPath string, handle RequestHandle) {
	log.Info("RegisterGetHandleNoUUID:", subPath)
	if subPath[0] != '/' {
		log.Panic("RegisterGetHandleNoUUID subPath must begin with '/', :", subPath)
	}

	path := rootPath + subPath
	h, _, _ := rootRouter.Lookup("GET", path)
	if h != nil {
		log.Panic("subPath with 'GET' has been register, subPath:", subPath)
	}

	rootRouter.GET(path, wrapGetHandleInternal(handle, false))
}
