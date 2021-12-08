#!/usr/bin/env bash

APP=rtsp2img
mkdir -p build/${APP}

export GOOS=windows
go build -o build/${APP}/${APP}.exe main.go
upx build/${APP}/${APP}.exe

export GOOS=linux
go build -o build/${APP}/${APP} main.go
upx build/${APP}/${APP}

cp ${APP}.json  build/${APP}/${APP}.json
tar cf build/${APP}.tar.gz build/${APP}