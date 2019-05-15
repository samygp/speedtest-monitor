#!/bin/bash
# go-speedtest

cd ..
set -e

echo -e "Go Building"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./docker/speedtest-monitor

cd docker/
echo -e "Making docker image"
docker build -t speedtest-monitor .