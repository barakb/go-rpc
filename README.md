go-rpc
======

[![Build Status](https://travis-ci.org/barakb/go-rpc.svg?branch=master)](https://travis-ci.org/barakb/go-rpc)


A very naive RPC implementation in GO.


* Without connection pool

Benchmark_ping-4	    5000	    268994 ns/op


* With connection pool

Benchmark_ping           	   50000	     23933 ns/op

Benchmark_ping-2         	   20000	     71243 ns/op

Benchmark_ping-4         	   20000	     83328 ns/op

Benchmark_ping_parallel  	   50000	     24326 ns/op

Benchmark_ping_parallel-2	  100000	     21652 ns/op

Benchmark_ping_parallel-4	  100000	     19437 ns/op

