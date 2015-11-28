package rpc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
)

type Connection struct {
	reader        *bufio.Reader
	writer        *bufio.Writer
	conn          net.Conn
	remoteAddress string
	localAddress  string
	context       *Context
	nextMessageId int
}

type Context struct {
	activeConnections map[string]*Connection
}

func NewContext() *Context {
	return &Context{make(map[string]*Connection)}
}

func (context *Context) Close() {
	for _, connection := range context.activeConnections {
		connection.Close()
	}
	context.activeConnections = make(map[string]*Connection)
}

func OpenConnection(remoteAddress string) (*Connection, error) {
	con, err := net.Dial("tcp", remoteAddress)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(con)
	writer := bufio.NewWriter(con)
	localAddress := con.LocalAddr().String()
	return &Connection{reader, writer, con, remoteAddress, localAddress, nil, 0}, nil
}

func Wrap(conn net.Conn, context *Context) *Connection {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	res := &Connection{reader, writer, conn, conn.LocalAddr().String(), conn.RemoteAddr().String(), context, 0}
	context.activeConnections[conn.LocalAddr().String()] = res
	return res
}

func (cc Connection) Close() error {
	if cc.context != nil {
		delete(cc.context.activeConnections, cc.conn.LocalAddr().String())
	}
	if err := cc.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Connection) NextMessageId() int {
	ret := c.nextMessageId
	c.nextMessageId++
	return ret
}

func (c Connection) LocalAddress() string {
	return c.localAddress
}

func (c Connection) RemoteAddress() string {
	return c.remoteAddress
}

func (cc Connection) Flush() error {
	return cc.writer.Flush()
}

func (c Connection) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c Connection) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

func (c Connection) readByte(data *byte) error {
	return binary.Read(c, binary.LittleEndian, data)
}

func (c Connection) readInt32(order binary.ByteOrder, data *int32) error {
	return binary.Read(c, order, data)
}

func (c Connection) String() string {
	return fmt.Sprintf("connection: %v <-> %v", c.localAddress, c.remoteAddress)
}
