// statserv
package main

import (
	"github.com/democratic-coin/dcoin-go/packages/stat"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"time"
	//	"fmt"
	"encoding/json"
	"log"
)

func daemon() {
	var (
		cur, max int64
		err      error
		iBalance *stat.InfoBalance
		pause    uint32 = 10
	)
	
	for {
		if cur == max {
			if max, err = utils.DB.Single(`select user_id from users order by user_id desc`).Int64(); err != nil {
				log.Println(`Error`, err)
			}
			cur = 1
			if err = stat.SetCashReqTime(); err != nil {
				log.Println(`Error`, err)
			}
			pause = GSettings.Period * 3600 / uint32(max)
			if pause == 0 {
				pause = 1
			}
			log.Println(`Start loop`, max, `/`, pause, `sec`)
//			max = 20
		}
		if iBalance, err = stat.GetBalance(cur); err != nil {
			log.Println(err)
		} else if len(iBalance.Currencies) > 0 {
			if idExist, err := GDB.Single(`select id from balance where user_id=? and date(uptime)=date('now')`,
				cur).Int64(); err == nil {
				if out, err := json.Marshal(iBalance); err == nil {
					if idExist > 0 {
						err = GDB.ExecSql(`update balance set data=?, uptime=datetime('now') where id=?`, out, idExist)
					} else {
						err = GDB.ExecSql(`insert into balance ( user_id, data, uptime) values( ?, ?,  datetime('now'))`,
							cur, out)
					}
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		}
		cur++
		time.Sleep(time.Duration(pause) * time.Second)
	}
}
