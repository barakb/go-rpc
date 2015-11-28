#!/usr/bin/env bash
go test  -bench BenchmarkNewConnectionPoolPingParallel -cpu=1,2,4,6,8,10,12,14,16,18,20 ./... -run XXX
