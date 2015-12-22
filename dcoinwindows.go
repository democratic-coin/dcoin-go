// +build windows

package main

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
	return C.w_ver()
}