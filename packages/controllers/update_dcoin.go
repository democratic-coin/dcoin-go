package controllers

import (
	"errors"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"github.com/c-darwin/dcoin-go/packages/consts"
)


func (c *Controller) UpdateDcoin() (string, error) {

	if c.SessRestricted != 0 || !c.NodeAdmin {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	_, url, err := utils.GetUpdVerAndUrl(consts.UPD_AND_VER_URL)
	if err!= nil {
		return "", utils.ErrInfo(err)
	}

	//fmt.Println(url)
	if len(url) > 0 {
		err = utils.DcoinUpd(url)
		if err!= nil {
			return "", utils.ErrInfo(err)
		}
		return utils.JsonAnswer("success", "success").String(), nil
	}
	return "", nil
}


