// +build windows

package main

import (
	"github.com/trayhost"
)

/*
#include <windows.h>
#include <stdio.h>
#include <stdlib.h>
int w_ver() {
	DWORD dwMajorVersion = 0;
	DWORD dwVersion = 0;
	dwVersion = GetVersion();
	dwMajorVersion = (DWORD)(LOBYTE(LOWORD(dwVersion)));
	return dwMajorVersion;
}*/
import "C"

func winVer() int {
	return int(C.w_ver())
}

func tray() {
	go func() {
		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://localhost:8089")
	}()
}

func enterLoop() {
	trayhost.EnterLoop("Dcoin", iconData)
}