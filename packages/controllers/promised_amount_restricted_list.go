package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

type promisedAmountRestrictedList struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	Pct float64
	Amount float64
	Lang            map[string]string
}

func (c *Controller) GetPromisedAmountCounter() ( float64, float64, error) {
	paRestricted, err := c.OneRow("SELECT * FROM promised_amount_restricted WHERE user_id = ?", c.SessUserId).String()
	if err != nil {
		return 0, 0, err
	}
	if _, ok := paRestricted[`user_id`]; !ok {
		return 0, 0, nil
	}
	
	amount := utils.StrToFloat64(paRestricted["amount"])
	profit, err := c.CalcProfitGen(utils.StrToInt64(paRestricted["currency_id"]), amount, c.SessUserId, utils.StrToInt64(paRestricted["start_time"]), utils.Time(), "wallet")
	if err != nil {
		return 0, 0, err
	}
	profit += amount

	pct, err := c.Single(c.FormatQuery("SELECT user FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC"), utils.StrToInt64(paRestricted["currency_id"])).Float64()
	if err != nil {
		return 0, 0, err
	}
	return profit, pct, nil
}

func (c *Controller) PromisedAmountRestrictedList() (string, error) {

	profit, pct, err := c.GetPromisedAmountCounter()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("promised_amount_restricted_list", "PromisedAmountRestrictedList", &promisedAmountRestrictedList{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		Pct : pct,
		Amount : profit,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
