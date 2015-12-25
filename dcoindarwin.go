// +build darwin
// +build 386 amd64

package main

import (
	//"time"
)
/*
#cgo darwin CFLAGS: -DDARWIN -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
void show ()
{
    [NSAutoreleasePool new];
    [NSApplication sharedApplication];
    [NSApp run];
}

*/
import "C"


func tray() {
	//C.show();
}

func enterLoop() {
	//time.Sleep(3600*24*90 * time.Second)
	C.show();
}