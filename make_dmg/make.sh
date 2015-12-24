diskutil unmount dcoin
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
rm packages/static/static.go
git stash
go get -u github.com/c-darwin/dcoin-go
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=amd64  CGO_ENABLED=1  go build -o make_dmg/dcoin.app/Contents/MacOs/dcoinbin
cd make_dmg
zip -r dcoin_osx64.zip dcoin.app/Contents/MacOs/dcoinbin
./make_dmg.sh -b background.png -i logo-big.icns -s "480:540" -c 240:400:240:200 -n dcoin_osx64 "dcoin.app"
cd ../
GOARCH=386  CGO_ENABLED=1  go build -o make_dmg/dcoin.app/Contents/MacOs/dcoinbin
cd make_dmg
zip -r dcoin_osx32.zip dcoin.app/Contents/MacOs/dcoinbin
diskutil unmount dcoin
mv dcoin.app/Contents/MacOS/ThrustShell .
./make_dmg.sh -b background.png -i logo-big.icns -s "480:540" -c 240:400:240:200 -n dcoin_osx32 "dcoin.app"
mv ThrustShell dcoin.app/Contents/MacOS/