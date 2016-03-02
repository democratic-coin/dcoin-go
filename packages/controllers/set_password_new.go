package controllers

import (
	"github.com/c-darwin/dcoin-go/packages/utils"
//	"html/template"
)

type setPasswordNewPage struct {
	Lang map[string]string
	IOS  bool
	Android  bool
	Mobile bool
	Community bool
	Email string
}

func (c *Controller) SetPasswordNew() (string, error) {
	var email string

	if c.Community {
		// если это пул, то будет прислан email
		email = c.Parameters["email"]
	}		
//	c.Lang[`need_key`] = template.HTML(c.Lang[`need_key`])
	TemplateStr, err := makeTemplate("set_password_new", "setPasswordNew", &setPasswordNewPage{
		Lang: c.Lang, IOS: utils.IOS(), Android: utils.Android(), Mobile: utils.Mobile(),
		Community: c.Community, Email: email })
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
