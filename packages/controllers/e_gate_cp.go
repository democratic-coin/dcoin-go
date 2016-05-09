package controllers

import (
	"fmt"
	//"github.com/democratic-coin/dcoin-go/packages/utils"
	//"errors"
)

func (c *Controller) EGateCP() (string, error) {

	c.r.ParseForm()

	fmt.Println(c.r.Form)
	log.Error("EGateCP %v", c.r.Form)

	fmt.Println(c.r.Header.Get("HTTP_HMAC"))
	log.Error("HTTP_HMAC %v", c.r.Header.Get("HTTP_HMAC"))

	fmt.Println(c.r.Header.Get("PHP_AUTH_USER"))
	log.Error("PHP_AUTH_USER %v", c.r.Header.Get("PHP_AUTH_USER"))

	fmt.Println(c.r.Header.Get("PHP_AUTH_PW"))
	log.Error("PHP_AUTH_PW %v", c.r.Header.Get("PHP_AUTH_PW"))

	for k, v := range c.r.Header {
		log.Error("key: %v / value: %v", k, v)
	}

/*
	currencyId := 0
	if c.r.FormValue("currency1") == "BTC" {
		currencyId = 1002
	}
	if currencyId == 0 {
		return "", errors.New("Incorrect currencyId")
	}

	amount := utils.StrToFloat64(c.r.FormValue("amount1"))
	pmId := utils.StrToInt64(c.r.FormValue("txn_id"))
	// проверим, не зачисляли ли мы уже это платеж
	existsId, err := c.Single(`SELECT id FROM e_adding_funds_cp WHERE id = ?`, pmId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if existsId != 0 {
		return "", errors.New("Incorrect txn_id")
	}
	paymentInfo := c.r.FormValue("item_name")

	txTime := utils.Time()
	err = EPayment(paymentInfo, currencyId, txTime, amount, pmId, "cp", c.ECommission)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
*/
	return ``, nil
}
