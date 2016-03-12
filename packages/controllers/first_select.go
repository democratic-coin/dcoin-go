package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

type firstSelectPage struct {
	Lang map[string]string
}

func (c *Controller) FirstSelect() (string, error) {

	TemplateStr, err := makeTemplate("first_select", "firstSelect", &delCreditPage{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
