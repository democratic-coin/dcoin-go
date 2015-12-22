package controllers

import (
	"errors"
	"github.com/c-darwin/dcoin-go/packages/consts"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"strings"
)

func (c *Controller) Update() (string, error) {

	if c.SessRestricted != 0 || !c.NodeAdmin {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	ver, _, err := utils.GetUpdVerAndUrl(consts.UPD_AND_VER_URL)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(ver) > 0 {
		newVersion := strings.Replace(c.Lang["new_version"], "[ver]", ver, -1)
		return utils.JsonAnswer(newVersion, "success").String(), nil
	}
	return "", nil
}
