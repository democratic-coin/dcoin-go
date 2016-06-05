// system
package system

import (
	"os"
//	"time"
	"github.com/go-thrust/thrust"
)

func Finish(exit int) {
	killChildProc()
//	time.Sleep(1*time.Second)
	if exit != 0 {
		os.Exit(exit)
	}
}

func FinishThrust(exit int) {
	thrust.Exit()
	Finish(exit)
}

