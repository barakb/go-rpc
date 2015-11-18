package rpc
import (
	"bufio"
	"net"
	"encoding/binary"
)


type Connection struct {
	reader *bufio.Reader
	writer *bufio.Writer
	conn   net.Conn
	address string;
}

func OpenConnection(address string) (*Connection, error) {
	con, err := net.Dial("tcp", address)
	if (err != nil) {
		return nil, err
	}
	reader := bufio.NewReader(con)
	writer := bufio.NewWriter(con)
	return &Connection{reader, writer, con, address}, nil
}

func wrap(conn  net.Conn)(*Connection){
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Connection{reader, writer, conn, ""}
}

func (cc Connection) Close() (error) {
	if err:= cc.conn.Close(); err != nil{
		return err
	}
	return nil
}

func (cc Connection) Flush() (error) {
	return cc.writer.Flush()
}

func (c Connection) Read(p []byte) (n int, err error){
	return c.reader.Read(p)
}

func (c Connection) Write(p []byte) (n int, err error){
	return c.writer.Write(p);
}


func (c Connection) readByte(data *byte) (error) {
	return binary.Read(c,binary.LittleEndian, data)
}

func (c Connection) readInt32(order binary.ByteOrder, data *int32) (error) {
	return binary.Read(c,order, data)
}

