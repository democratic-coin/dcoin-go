#!/bin/bash
if [ $# -gt 0 ] && [ $1 = "debug" ] 
then
  DEBUG="-debug=true"
fi
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" $DEBUG static/...
