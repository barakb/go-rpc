package pool

import (
	"fmt"
	"net"
	"testing"
)

func TestConnectionPool(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Errorf("Error listening: %v", err.Error())
	}
	defer l.Close()
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Errorf("Error accepting: %v", err.Error())
				return
			}
			fmt.Printf("conn accepted %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())
		}
	}()

	address := l.Addr().String()
	cp := NewConnectionPool(2)
	for i := 0; i < 3; i++ {
		con, err := cp.Get(address)
		if err != nil {
			t.Errorf("Fail to get connection %v", err)
		}
		cp.Put(con)
	}
}
