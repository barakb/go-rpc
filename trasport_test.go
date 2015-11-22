package rpc

import (
	"fmt"
	"testing"
	"time"
)

func TestTrasport_StartStop(t *testing.T) {
	trasport := NewTCPTransport("127.0.0.1:0", time.Second, nil)
	go func() {
		// this is the server side
		// it should read message from the consumer channel and reply to them.
		rpc := <-trasport.Consumer()
		fmt.Printf("consumer got rpc command %#v\n", rpc.Command)
		rpc.Respond(&EchoResponse{"FOO"}, nil)
	}()

	// this is the client side
	res, err := trasport.Echo(trasport.LocalAddr(), "foo")
	if err != nil {
		t.Errorf("error %#v\n", err)
	}
	if res != "FOO" {
		t.Errorf("expected FOO instead %v", res)
	}
	fmt.Printf("returns %v\n", res)
}

func TestTrasport_StartClose(t *testing.T) {
	trasport := NewTCPTransport("127.0.0.1:0", time.Second, nil)
	go func() {
		<-trasport.Consumer()
		trasport.Close()
	}()

	// this is the client side
	_, err := trasport.Echo(trasport.LocalAddr(), "foo")
	if err == nil {
		t.Errorf("Expected error because trasport is closed")
	}
}

func Benchmark_ping(b *testing.B) {
	trasport := NewTCPTransport("127.0.0.1:0", time.Second, NewEmptyLogger())
	go func() {
		for {
			// this is the server side
			// it should read message from the consumer channel and reply to them.
			rpc := <-trasport.Consumer()
			rpc.Respond(&EchoResponse{"FOO"}, nil)
		}
	}()

	for i := 0; i < b.N; i++ {
		res, err := trasport.Echo(trasport.LocalAddr(), "foo")
		if err != nil {
			b.Errorf("expected FOO instead error %v", err)
		}
		trasport.Info("res is: ", res)

	}
}

func Benchmark_ping_parallel(b *testing.B) {
	trasport := NewTCPTransport("127.0.0.1:0", time.Second, NewEmptyLogger())
	go func() {
		for {
			// this is the server side
			// it should read message from the consumer channel and reply to them.
			rpc := <-trasport.Consumer()
			rpc.Respond(&EchoResponse{"FOO"}, nil)
		}
	}()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := trasport.Echo(trasport.LocalAddr(), "foo")
			if err != nil {
				b.Errorf("expected FOO instead error %v", err)
			}
			trasport.Info("res is: ", res)
		}
	})
}
