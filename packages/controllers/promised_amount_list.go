package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"strings"
	"time"
)

type promisedAmountListPage struct {
	Alert                        string
	SignData                     string
	ShowSignData                 bool
	TxType                       string
	TxTypeId                     int64
	TimeNow                      int64
	UserId                       int64
	Lang                         map[string]string
	CountSignArr                 []int
	LastTxFormatted              string
	CurrencyList                 map[int64]string
	ConfigCommission             map[int64][]float64
	Navigate                     string
	Commission                   map[int64][]float64
	PromisedAmountListAccepted   []utils.PromisedAmounts
	ActualizationPromisedAmounts int64
	LimitsText                   string
}

func (c *Controller) PromisedAmountList() (string, error) {

	txType := "PromisedAmount"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewPromisedAmount", "ChangePromisedAmount", "DelPromisedAmount", "ForRepaidFix", "ActualizationPromisedAmounts", "Mining"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}

	limitsText := strings.Replace(c.Lang["change_commission_limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_promised_amount"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_promised_amount_period"]], -1)

	actualizationPromisedAmounts, promisedAmountListAccepted, _, err := c.GetPromisedAmounts(c.SessUserId, c.Variables.Int64["cash_request_time"])
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, tx := range last_tx {
		if utils.StrToInt64( tx[`block_id`] ) == 0 {
			IDB:
			for _, idb := range []string{`queue_tx`,`transactions`}{
				data, err := c.Single(`SELECT data FROM `+ idb +` WHERE hex(hash)=?`, utils.BinToHex([]byte(tx["hash"])) ).Bytes(); 
				if len( data ) > 0 {
					data2 := data[5:]			
					length := utils.DecodeLength(&data2)
					utils.BytesShift(&data2, length)
					length = utils.DecodeLength(&data2)
					idPromise := utils.StrToInt64( string(utils.BytesShift(&data2, length)))
					for i, ipromise := range promisedAmountListAccepted {
						if ipromise.Id == idPromise {
							promisedAmountListAccepted[i].InProcess = true
							break IDB
						}
					}
				}
			}
		}
	}
	TemplateStr, err := makeTemplate("promised_amount_list", "promisedAmountList", &promisedAmountListPage{
		Alert:                        c.Alert,
		Lang:                         c.Lang,
		CountSignArr:                 c.CountSignArr,
		ShowSignData:                 c.ShowSignData,
		UserId:                       c.SessUserId,
		TimeNow:                      timeNow,
		TxType:                       txType,
		TxTypeId:                     txTypeId,
		SignData:                     "",
		LastTxFormatted:              lastTxFormatted,
		CurrencyList:                 c.CurrencyList,
		PromisedAmountListAccepted:   promisedAmountListAccepted,
		ActualizationPromisedAmounts: actualizationPromisedAmounts,
		LimitsText:                   limitsText})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
