#! /bin/bash -e
./bindata.sh
#go get -u github.com/democratic-coin/dcoin-go
GOARCH=386  CGO_ENABLED=1  go build -o make_deb/dcoin/usr/share/dcoin/dcoin
GOARCH=amd64  CGO_ENABLED=1  go build -o make_deb/dcoin64/usr/share/dcoin/dcoin
cd make_deb
