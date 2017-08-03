#!/bin/bash


export GOOS=linux
export GOARCH=amd64
export VERSION=0.1.1
go build *.go
mv dgraph_helper dgraph_helper_v${VERSION}_${GOOS}_${GOARCH}
chmod +x dgraph_helper_v${VERSION}_${GOOS}_${GOARCH}
echo "BUILT dgraph_helper_v${VERSION}_${GOOS}_${GOARCH}"
