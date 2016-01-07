// +build android

package dcoin

import (
	"net/http"
	"github.com/c-darwin/dcoin-go/packages/utils"
)

func IosLog(text string) {
}

func KillPid(pid string) error {
	return nil
}

func httpListener(ListenHttpHost, BrowserHttpHost string) {
	go func() {
		http.ListenAndServe(ListenHttpHost, nil)
	}()
}

func tcpListener() {

}

func tray() {

}

func signals(chans []*utils.DaemonsChans) {

}
