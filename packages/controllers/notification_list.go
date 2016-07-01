package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

type notificationListPage struct {
	Lang            map[string]string
	LangInt         int64
	List            []map[string]string
}

func (c *Controller) NotificationList() (string, error) {

	list, err := c.GetAll("SELECT * FROM notifications WHERE user_id = ? AND isread=1 ORDER BY id DESC", 30, c.SessUserId )
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	
	TemplateStr, err := makeTemplate("notification_list", "notification_list", &notificationListPage{
		Lang:            c.Lang,
		LangInt:         c.LangInt,
		List:            list})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
