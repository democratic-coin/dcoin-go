// stat
package stat

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"github.com/democratic-coin/dcoin-go/packages/controllers"
	"time"
//	"encoding/json"
/*	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fmt"*/
)

const (
	STAT_SERVER = `http://localhost:8091`
//	STAT_SERVER = `http://stat.dcoin.club:8201`
)

type CurrencyBalance struct {
	CurrencyId int64    `json:"cur_id"`
	Wallet     float64  `json:"wallet"`
	Tdc        float64  `json:"tdc"` 
	Promised   float64  `json:"promised"`
	Restricted float64  `json:"restricted"`
	Summary    float64  `json:"summary"`
}

type InfoBalance struct {
	Currencies map[string] *CurrencyBalance
	Time       int64
}

type HistoryBalance struct {
	Success  bool            `json:"success"`
	Error    string          `json:"error"`
	History  []*InfoBalance  `json:"history"`
}

var (
	cashReqTime  int64
)

func SetCashReqTime() error {
	if vars, err := utils.DB.GetAllVariables(); err != nil {
		return err
	} else {
		cashReqTime = vars.Int64["cash_request_time"]
	}
	return nil
}

func GetBalance(userId int64) (*InfoBalance,error) {
	
	ret := new(InfoBalance)		
	list := make(map[string]*CurrencyBalance)

	if wallet, err := utils.DB.GetBalances(userId); err == nil {
		for _, iwallet := range wallet {
			list[utils.Int64ToStr(iwallet.CurrencyId)] = &CurrencyBalance{ CurrencyId: iwallet.CurrencyId,
			                Wallet: utils.Round(iwallet.Amount, 6) }
		}
	} else {
		return ret, err
	}
	if cashReqTime == 0 {
		if err := SetCashReqTime(); err!=nil {
			return ret, err
		}
	}
	if _, dc, _, err := utils.DB.GetPromisedAmounts(userId, cashReqTime); err == nil {
		for _, idc := range dc {
			currency := utils.Int64ToStr(idc.CurrencyId)
			if _, ok:= ret.Currencies[currency]; ok {
				list[currency].Tdc += utils.Round(idc.Tdc,6)
				list[currency].Promised += idc.Amount
			} else {
				list[currency] = &CurrencyBalance{ CurrencyId: idc.CurrencyId,
			                Promised: idc.Amount, Tdc: utils.Round(idc.Tdc, 6) }
				}
			} 
		} else {
			return ret,err
		}

	c := new(controllers.Controller)
	c.SessUserId = userId
	c.DCDB = utils.DB
	if profit,_, err := c.GetPromisedAmountCounter(); err == nil && profit > 0 {
		currency := `72`
		if _, ok:= list[currency]; ok {
			list[currency].Restricted = utils.Round( profit - 30, 6)
		} else {
			list[currency] = &CurrencyBalance{ CurrencyId: utils.StrToInt64(currency),
			                Restricted: utils.Round( profit - 30, 6) }
		}
	}
	for i := range list {
		list[i].Summary = utils.Round( list[i].Wallet + list[i].Tdc + list[i].Restricted, 6 )
	}
	ret.Currencies = list
	ret.Time = time.Now().Unix()
	return ret,nil
}

