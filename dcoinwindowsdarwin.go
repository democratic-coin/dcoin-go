// +build windows darwin
// +build 386 amd64

package main

import (
	"github.com/c-darwin/trayhost"
)

func tray() {
	go func() {
		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://localhost:8089")
	}()
}

func EnterLoop() {
	trayhost.EnterLoop("Dcoin", iconData)
}