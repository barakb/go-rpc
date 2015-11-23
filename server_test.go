package rpc

import (
	"fmt"
	"testing"
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

func TestServer_StartClose(t *testing.T) {
	server := NewServer(nil)
	go func() {
		<-server.Consumer()
		server.Close()
	}()

	// this is the client side
	_, err := server.Echo(server.LocalAddr(), "foo")
	if err == nil {
		t.Errorf("Expected error because trasport is closed")
	}
}

func Benchmark_ping(b *testing.B) {
	server := NewServer(NewEmptyLogger())
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

func Benchmark_ping_parallel(b *testing.B) {
	server := NewServer(NewEmptyLogger())
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
