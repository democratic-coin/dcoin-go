#! /bin/bash -e
ARCH0=""
if [ $# -gt 0 ] && [ $1 = "amd64" ]
then
  ARCH0="64"
fi
./bindata.sh
#go get -u github.com/democratic-coin/dcoin-go
GOARCH=$1  CGO_ENABLED=1  go build -o make_deb/dcoin$ARCH0/usr/share/dcoin/dcoin
cd make_deb