// +build !android,!ios

package main

import (
	"github.com/c-darwin/dcoin-go/packages/dcoin"
)

func main() {
	dcoin.Start("")
}