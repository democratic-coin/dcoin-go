package controllers

import (
	"github.com/c-darwin/dcoin-go/packages/utils"
	"github.com/c-darwin/dcoin-go/packages/daemons"
	"regexp"
)

func (c *Controller) RestartDb() (string, error) {

	if ok, _ := regexp.MatchString(`(\:\:)|(127\.0\.0\.1)`, c.r.RemoteAddr); ok {
		err := daemons.ClearDb(nil, "")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else {
		return "", utils.ErrInfo("Access denied for "+c.r.RemoteAddr)
	}
	return "", nil
}
