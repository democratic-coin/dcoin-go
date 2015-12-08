// +build linux darwin freebsd
// +build 386 amd64

package dcoin

import (
	"syscall"
	"github.com/c-darwin/dcoin-go/packages/utils"
)

func KillPid(pid string) error {
	err := syscall.Kill(utils.StrToInt(pid), syscall.SIGTERM)
	if err != nil {
		return err
	}
	return nil
}