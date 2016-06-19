// statserv
package main

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"github.com/democratic-coin/dcoin-go/packages/controllers"
	"time"
//	"fmt"
	"log"
	"encoding/json"
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

var (
	cashReqTime  int64
)

func getBalance(userId int64) (*InfoBalance,error) {
	
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

func daemon() {
	var (
		cur,max int64
		err     error
		iBalance *InfoBalance
	)
	pause := 10
	for {
		if cur == max {
			if max, err = utils.DB.Single(`select user_id from users order by user_id desc`).Int64(); err != nil {
 				log.Println(`Error`, err )
			}
			cur = 1
			if vars, err := utils.DB.GetAllVariables(); err == nil {
				cashReqTime = vars.Int64["cash_request_time"]
			} else {
				log.Println(`Error`, err )
			}
			pause = 23*3600/int(max)
			if pause == 0 {
				pause = 1
			}
			log.Println(`Start loop`, max, `/`, pause, `sec`)
//			max = 20
		}
		if iBalance, err = getBalance( cur ); err != nil {
			log.Println( err )
		} else {
			if idExist,err := GDB.Single(`select id from balance where user_id=? and date(uptime)=date('now')`, 
			                 cur ).Int64(); err == nil {
				if out,err := json.Marshal( iBalance ); err == nil {
					if idExist > 0 {
						err = GDB.ExecSql(`update balance set data=?, uptime=datetime('now') where id=?`, out, idExist )
					} else {
						err = GDB.ExecSql(`insert into balance ( user_id, data, uptime) values( ?, ?,  datetime('now'))`,
			                  cur, out )
					}
					if err != nil {
						log.Println( err )
					}
				} else {
					log.Println( err )
				}
			} else {
				log.Println( err )
			}
		}
		cur++
		time.Sleep( time.Duration(pause) * time.Second )
	}
}
