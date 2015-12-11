go get -u github.com/jteeuwen/go-bindata/... 
cd %GOPATH/%src/github.com/c-darwin/dcoin-go
git stash
go get -u github.com/c-darwin/dcoin-go 
go-bindata -o="packages/static/static.go" -pkg="static" static/... 
go install -ldflags "-H windowsgui" github.com/c-darwin/dcoin-go
mv C:\go-projects\bin\dcoin-go.exe C:\go-projects\bin\dcoin.exe
pause