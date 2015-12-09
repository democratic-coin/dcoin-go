#! /bin/bash -e

#go get -u github.com/c-darwin/dcoin-go
GOARCH=386  CGO_ENABLED=1  go build -o make_deb/dcoin/usr/share/dcoin/dcoin
GOARCH=amd64  CGO_ENABLED=1  go build -o make_deb/dcoin64/usr/share/dcoin/dcoin
cd make_deb
dpkg-deb --build dcoin
dpkg-deb --build dcoin64

zip -j dcoin_linux32.zip dcoin/usr/share/dcoin/dcoin
zip -j dcoin_linux64.zip dcoin64/usr/share/dcoin/dcoin
mv dcoin_linux32.zip /home/z/multiplatform/dc-compiled/dcoin_linux32.zip
mv dcoin_linux64.zip /home/z/multiplatform/dc-compiled/dcoin_linux64.zip
mv dcoin.deb /home/z/multiplatform/dc-compiled/dcoin_linux32.deb
mv dcoin64.deb /home/z/multiplatform/dc-compiled/dcoin_linux64.deb
rm -rf dcoin64/usr/share/dcoin/dcoin
rm -rf dcoin/usr/share/dcoin/dcoin
