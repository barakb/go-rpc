package rpc

import (
	"bufio"
	"encoding/binary"
	"net"
)

type Connection struct {
	reader  *bufio.Reader
	writer  *bufio.Writer
	conn    net.Conn
	address string
	context *Context
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

func OpenConnection(address string) (*Connection, error) {
	con, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(con)
	writer := bufio.NewWriter(con)
	return &Connection{reader, writer, con, address, nil}, nil
}

func wrap(conn net.Conn, context *Context) *Connection {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	res := &Connection{reader, writer, conn, "", context}
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
