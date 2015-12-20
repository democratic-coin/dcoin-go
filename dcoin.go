// +build !android,!ios

package main

import (
	"github.com/c-darwin/dcoin-go/packages/dcoin"
	"github.com/c-darwin/trayhost"
	"runtime"

)

func main() {

	runtime.LockOSThread()

	go func() {
		// Run your application/server code in here. Most likely you will
		// want to start an HTTP server that the user can hit with a browser
		// by clicking the tray icon.

		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://localhost:8089")
		trayhost.EnterLoop("Dcoin", iconData)
	}()

	dcoin.Start("")


}