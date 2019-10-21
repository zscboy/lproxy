package sayhello
import (
	log "github.com/sirupsen/logrus"
	context "context"
	"lproxy/server"
)

type myGreeterServer struct {
}

func (s *myGreeterServer) SayHello(c context.Context, r *HelloRequest) (*HelloReply, error) {
	log.Println("request name:", r.GetName())

	reply := &HelloReply{Message: "hello boy"}
	return reply, nil
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		s := server.GetGRPCServer()
		RegisterGreeterServer(s, &myGreeterServer{})
	})
}
