package pool

import (
	"fmt"
	"testing"
	"com.githab.barakb/rpc"
)

func TestServer_StartStop(t *testing.T) {
	server := NewServer(nil)
	go func() {
		// this is the server side
		// it should read message from the consumer channel and reply to them.
		rpc := <-server.Consumer()
		fmt.Printf("consumer got rpc command %#v\n", rpc.Command)
		rpc.Respond(&EchoResponse{"FOO"}, nil)
	}()

	// this is the client side
	res, err := server.Echo(server.LocalAddr(), "foo")
	if err != nil {
		t.Errorf("error %#v\n", err)
	}
	if res != "FOO" {
		t.Errorf("expected FOO instead %v", res)
	}
	fmt.Printf("returns %v\n", res)
}


func BenchmarkNewConnectionPoolPing(b *testing.B) {
	server := NewServer(rpc.NewEmptyLogger())
	go func() {
		for {
			// this is the server side
			// it should read message from the consumer channel and reply to them.
			rpc := <-server.Consumer()
			rpc.Respond(&EchoResponse{"FOO"}, nil)
		}
	}()

	for i := 0; i < b.N; i++ {
		res, err := server.Echo(server.LocalAddr(), "foo")
		if err != nil {
			b.Errorf("expected FOO instead error %v", err)
		}
		server.Info("res is: ", res)

	}
}

func BenchmarkNewConnectionPoolPingParallel(b *testing.B) {
	server := NewServer(rpc.NewEmptyLogger())
	go func() {
		for {
			// this is the server side
			// it should read message from the consumer channel and reply to them.
			rpc := <-server.Consumer()
			rpc.Respond(&EchoResponse{"FOO"}, nil)
		}
	}()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := server.Echo(server.LocalAddr(), "foo")
			if err != nil {
				b.Errorf("expected FOO instead error %v", err)
			}
			server.Info("res is: ", res)
		}
	})
}
