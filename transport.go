package rpc

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type tcpTransport struct {
	Logger
	bindAddr      string
	listenAddress net.Addr
	timeout       time.Duration
	consumer      chan RPC
}

func NewTCPTransport(bindAddr string, timeout time.Duration, logger Logger) *tcpTransport {
	if logger == nil {
		logger = NewLogger(os.Stdout)
	}
	res := &tcpTransport{Logger: logger, bindAddr: bindAddr, timeout: timeout,
		consumer: make(chan RPC)}
	addressChannel := make(chan net.Addr)
	go res.listen(addressChannel)
	// do not return before the server publish itself.
	res.listenAddress = <-addressChannel
	close(addressChannel)
	return res
}

type EchoRequest struct {
	Msg string
}

type EchoResponse struct {
	Msg string
}

func (t *tcpTransport) LocalAddr() string {
	if (t.listenAddress != nil) {
		return t.listenAddress.String()
	}
	return t.bindAddr
}

func (t *tcpTransport) Consumer() <-chan RPC {
	return t.consumer
}

func (t *tcpTransport) Echo(target string, msg string) (string, error) {
	t.Debug("Echo to  %s\n", target)
	req := &EchoRequest{msg}
	resp := &EchoResponse{}
	if err := genericRPC(target, 0, req, resp); err != nil {
		return "", err
	}
	return resp.Msg, nil
}

func (t *tcpTransport) listen(addressChannel chan net.Addr) {
	server, err := net.Listen("tcp", t.bindAddr)
	if server == nil {
		t.Info("couldn't start listening: %v\n", err)
	}
	addressChannel <- server.Addr()
	t.Debug("Starting listener at %s\n", server.Addr().String())
	for {
		connection, err := server.Accept()
		if err != nil {
			t.Info("couldn't accept connection : %v\n", err)
			continue
		}
		go t.handleConnection(connection)
	}
}

func (t *tcpTransport) handleConnection(conn net.Conn) {
	t.Debug("handleConnection: %v <-> %v\n", conn.LocalAddr(), conn.RemoteAddr())
	for {
		var rpcType byte
		err := binary.Read(conn, binary.LittleEndian, &rpcType)
		if err != nil {
			if err != io.EOF {
				t.Debug("failed to read rpcType  %v <-> %v error is %#v\n", conn.LocalAddr(), conn.RemoteAddr(), err)
			}
			conn.Close()
			return
		}

		var size int32
		err = binary.Read(conn, binary.LittleEndian, &size)
		if err != nil {
			t.Debug("binary.Read failed:", err)
		}

		buf := make([]byte, size)

		conn.Read(buf)
		req := &EchoRequest{}
		if err := json.Unmarshal(buf, &req); err != nil {
			t.Debug("failed to Unmarshal request from client %s\n", conn.RemoteAddr())
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
			t.Debug("failed to reply %#v on message %#v to client %s\n", resp, req, conn.RemoteAddr())
			conn.Close()
			return
		}
	}
}

func genericRPC(address string, rpcType uint8, args interface{}, resp interface{}) error {
	conn, err := openConnection(address)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open client connection to %s for sending request %#v error is %v.", address, args, err))
	}

	defer conn.Close()

	if err := sendRPC(conn, rpcType, args); err != nil {
		return err
	}
	return decodeResponse(conn, resp)
}

func sendRPC(conn net.Conn, rpcType uint8, args interface{}) error {
	// write message type
	if err := binary.Write(conn, binary.LittleEndian, byte(0)); err != nil {
		return err
	}

	// marshal args
	bytes, err := json.Marshal(args)

	if err != nil {
		return err
	}

	// write args len as int32
	if err := binary.Write(conn, binary.LittleEndian, int32(len(bytes))); err != nil {
		return err
	}

	// write marshalled args
	if _, err = conn.Write(bytes); err != nil {
		return err
	}

	return nil
}

func (t *tcpTransport) sendReplyFromServer(conn net.Conn, response *RPCResponse) error {
	t.Debug("sendReplyFromServer %v -> %v %#v \n", conn.LocalAddr(), conn.RemoteAddr(), response.Response)
	var bytes []byte
	var err error
	if response.Error != nil {
		if err := binary.Write(conn, binary.LittleEndian, byte(1)); err != nil {
			return err
		}
		// marshal args
		bytes, err = json.Marshal(response.Error.Error())
		if err != nil {
			return err
		}
	} else {
		if err := binary.Write(conn, binary.LittleEndian, byte(0)); err != nil {
			return err
		}
		// marshal args
		bytes, err = json.Marshal(response.Response)
		if err != nil {
			return err
		}

	}

	// write args len as int32
	if err := binary.Write(conn, binary.LittleEndian, int32(len(bytes))); err != nil {
		return err
	}
	// write marshalled args
	if _, err = conn.Write(bytes); err != nil {
		return err
	}


	return nil
}

func decodeResponse(conn net.Conn, resp interface{}) error {
	var isError byte
	if err := binary.Read(conn, binary.LittleEndian, &isError); err != nil {
		return err
	}

	var size int32
	err := binary.Read(conn, binary.LittleEndian, &size)
	if err != nil {
		return err
	}
	buf := make([]byte, size)

	_, err = conn.Read(buf)
	if err != nil {
		return err
	}
	if isError == 1 {
		var errorMessage string
		if err := json.Unmarshal(buf, &errorMessage); err != nil {
			return err
		}
		return errors.New(errorMessage)
	} else {
		if err := json.Unmarshal(buf, resp); err != nil {
			return err
		}
	}
	return nil
}

func openConnection(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
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
