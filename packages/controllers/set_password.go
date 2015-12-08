package controllers

import (
	"github.com/c-darwin/dcoin-go/packages/utils"
)

type setPasswordPage struct {
	Lang map[string]string
	IOS  bool
}

func (c *Controller) SetPassword() (string, error) {

	var ios bool
	if utils.IOS() {
		ios = true
	}
	TemplateStr, err := makeTemplate("set_password", "setPassword", &setPasswordPage{
		Lang: c.Lang, IOS: ios})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
