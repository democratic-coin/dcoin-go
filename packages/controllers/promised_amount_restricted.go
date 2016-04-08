package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

type promisedAmountRestricted struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	TxType           string
	TxTypeId         int64
	TimeNow          int64
	Lang            map[string]string
}

func (c *Controller) PromisedAmountRestricted() (string, error) {

	txType := "NewRestrictedPromisedAmount"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("promised_amount_restricted", "PromisedAmountRestricted", &promisedAmountRestricted{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		TimeNow:          timeNow,
		TxType:           txType,
		TxTypeId:         txTypeId,
		UserId:          c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
