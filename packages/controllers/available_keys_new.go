package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/availablekey"
	"github.com/democratic-coin/dcoin-go/packages/utils"
//	"fmt"
)

type availableKeysNewPage struct {
	AutoLogin bool
	Key       string
	LangId    int
}

func (c *Controller) AvailableKeysNew() (string, error) {

	var email string
	if c.Community {
		// если это пул, то будет прислан email
		email = c.r.FormValue("email")
		if !utils.ValidateEmail(email) {
			return utils.JsonAnswer("Incorrect email", "error").String(), nil
		}
		// если мест в пуле нет, то просто запишем юзера в очередь
		pool_max_users, err := c.Single("SELECT pool_max_users FROM config").Int()
		if err != nil {
			return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
		}
		if len(c.CommunityUsers) >= pool_max_users {
			err = c.ExecSql("INSERT INTO pool_waiting_list ( email, time, user_id ) VALUES ( ?, ?, ? )", email, utils.Time(), 0)
			if err != nil {
				return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
			}
			return utils.JsonAnswer(c.Lang["pool_is_full"], "error").String(), nil
		}
	}

	availablekey := &availablekey.AvailablekeyStruct{}
	availablekey.DCDB = c.DCDB
	availablekey.Email = email
	userId, publicKey, err := availablekey.GetAvailableKey()
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	if userId > 0 {
		c.sess.Set("user_id", userId)
		c.sess.Set("public_key", publicKey)
		log.Debug("user_id: %d", userId)
		log.Debug("public_key: %s", publicKey)
		return utils.JsonAnswer("success", "success").String(), nil
	}
	return utils.JsonAnswer("no_available_keys", "error").String(), nil
}
