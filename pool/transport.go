package pool


import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
	"github.com/barakb/rpc"
)

var marshaller *rpc.Marshaller

type tcpTransport struct {
	rpc.Logger
	bindAddr       string
	listenAddress  net.Addr
	timeout        time.Duration
	consumer       chan RPC
	connectionPool *ConnectionPool
	server         net.Listener
	context        *rpc.Context
	quit           chan interface{}
}

func NewTCPTransport(bindAddr string, timeout time.Duration, connectionPool *ConnectionPool, logger rpc.Logger) *tcpTransport {
	if logger == nil {
		logger = rpc.NewLogger(os.Stdout)
	}
	res := &tcpTransport{Logger: logger, bindAddr: bindAddr, timeout: timeout,
		consumer: make(chan RPC), connectionPool: connectionPool,
		context: rpc.NewContext()}
	res.quit = make(chan interface{})
	addressChannel := make(chan net.Addr)
	go res.listen(addressChannel)
	// do not return before the server publish itself.
	res.listenAddress = <-addressChannel
	close(addressChannel)
	return res
}


func (t *tcpTransport) LocalAddr() string {
	if t.listenAddress != nil {
		return t.listenAddress.String()
	}
	return t.bindAddr
}

func (t *tcpTransport) Consumer() <-chan RPC {
	return t.consumer
}

func (t *tcpTransport) Close() {
	//	close(t.quit)
	//	t.server.Close()
	//	t.connectionPool.Close()
	//	t.context.Close()
}

func (t *tcpTransport) listen(addressChannel chan net.Addr) {
	var err error
	t.server, err = net.Listen("tcp", t.bindAddr)
	if t.server == nil {
		t.Info("couldn't start listening: %v\n", err)
	}
	addressChannel <- t.server.Addr()
	t.Debug("Starting listener at %s\n", t.server.Addr().String())
	for {
		connection, err := t.server.Accept()
		if err != nil {
			select {
			case <-t.quit:
				return
			default:
				t.Info("Failed to accept connection : %#v \n", err)
				continue
			}
		}
		con := rpc.Wrap(connection, t.context)
		go t.handleConnection(con)
	}
}

func (t *tcpTransport) handleConnection(conn *rpc.Connection) {
	t.Debug("handleConnection: %v\n", *conn)
	for {
		req, err := marshaller.UnMarshalRequest(conn)
		if err != nil {
			if err != io.EOF {
				t.Debug("failed to read rpcType  %v, error is %#v\n", *conn, err)
			}
			conn.Close()
			return
		}

		respCh := make(chan RPCResponse, 1)
		rpc := RPC{
			RespChan: respCh,
		}
		rpc.Command = req

		t.consumer <- rpc

		t.Debug("Sending command %#v to consumer, waiting for consumer response\n", rpc.Command)
		resp := <-respCh
		t.Debug("server got consumer respond %#v, sending it back to client\n", resp)

		if err := t.sendReplyFromServer(conn, &resp); err != nil {
			t.Debug("failed to reply %#v on message %#v to client %s\n", resp, req, conn.RemoteAddress())
			conn.Close()
			return
		}
	}
}

func (t *tcpTransport) genericRPC(address string, rpcType uint8, args interface{}, resp interface{}) error {
	//	conn, err := openConnection(address)
	conn, err := t.connectionPool.Get(address)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open client connection to %s for sending request %#v error is %v.", address, args, err))
	}

	//	defer conn.Close()
	defer t.connectionPool.Put(conn)

	if err := sendRPC(conn, rpcType, args); err != nil {
		return err
	}
	return marshaller.UnMarshalResponse(conn, resp)
}

func sendRPC(conn *rpc.Connection, rpcType uint8, args interface{}) error {
	if err := marshaller.Marshal(conn, byte(0), args); err != nil {
		return err
	}
	conn.Flush()
	return nil
}

func (t *tcpTransport) sendReplyFromServer(conn *rpc.Connection, response *RPCResponse) error {
	t.Debug("sendReplyFromServer %v -> %v %#v \n", conn.LocalAddress(), conn.RemoteAddress(), response.Response)
	if response.Error != nil {
		if err := marshaller.Marshal(conn, byte(1), response.Error.Error()); err != nil {
			return err
		}
	} else {
		if err := marshaller.Marshal(conn, byte(0), response.Response); err != nil {
			return err
		}
	}
	conn.Flush()
	return nil
}

type RPCResponse struct {
	Response interface{}
	Error    error
}

// RPC has a command, and provides a response mechanism.
type RPC struct {
	Command  interface{}
	RespChan chan <- RPCResponse
}

// Respond is used to respond with a response, error or both
func (r *RPC) Respond(resp interface{}, err error) {
	r.RespChan <- RPCResponse{resp, err}
}

type Transport interface {
	Consumer() <-chan RPC
	LocalAddr() string
	Echo(target string, msg string) (string, error)
}

