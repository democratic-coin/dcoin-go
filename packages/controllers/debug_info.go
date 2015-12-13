package controllers

import (
	"github.com/c-darwin/dcoin-go/packages/utils"
	"encoding/json"
	"github.com/c-darwin/dcoin-go/packages/consts"
)

func (c *Controller) DebugInfo() (string, error) {

	mainLock, err := c.OneRow(`SELECT * FROM main_lock`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	jsonMainLock, err := json.Marshal(mainLock)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	nodesBan, err := c.GetAll(`SELECT * FROM nodes_ban`, 20)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	jsonNodesBan, err := json.Marshal(nodesBan)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return consts.VERSION+"\n"+string(jsonMainLock)+"\n"+string(jsonNodesBan)+"\n", nil
}
