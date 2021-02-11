#!/bin/bash
redis-cli --version
/usr/bin/redis-server --daemonize yes
go test -v ./tests/...