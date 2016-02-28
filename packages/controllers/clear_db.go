package controllers

import (
	"errors"
	"src/github.com/astaxie/beego/config"
	"github.com/c-darwin/dcoin-go/packages/utils"
)

func (c *Controller) ClearDb() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	err := c.ExecSql(`UPDATE install SET progress = 0`)
	if err != nil {
		utils.Mutex.Unlock()
		return "", utils.ErrInfo(err)
	}

	confIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	confIni.Set("db_type", "")
	err = confIni.SaveConfigFile(*utils.Dir + "/config.ini")

	return "", nil
}
