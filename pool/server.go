package pool
import (
"time"
"github.com/barakb/rpc"
)

type Server struct {
	*tcpTransport
}

func NewServer(logger rpc.Logger) *Server{
	return &Server{NewTCPTransport(":0", time.Second, NewConnectionPool(4), logger)}
}

type EchoRequest struct {
	Msg string
}

type EchoResponse struct {
	Msg string
}


func (s *Server) Echo(target string, msg string) (string, error) {
	s.tcpTransport.Debug("Echo to  %s\n", target)
	req := &EchoRequest{msg}
	resp := &EchoResponse{}
	if err := s.tcpTransport.genericRPC(target, 0, req, resp); err != nil {
		return "", err
	}
	return resp.Msg, nil
}

//func (s *Server) Consumer() <-chan RPC {
//	return s.tcpTransport.consumer
//}

