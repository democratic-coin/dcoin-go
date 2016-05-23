// emailserv
package main

import (
	"net/http"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"github.com/democratic-coin/dcoin-go/packages/controllers"
//	"html/template"
	"bytes"
	"hash/crc32"
	"strconv"
//	"io"
	"fmt"
	"strings"
	"sync"
)

type balanceTask struct {
	UserId  int64
	Error   error
}

var (
	queueBalance  []*balanceTask
	bCurrent  int   
	bMutex    sync.Mutex
)

func init() {
	queueBalance = make([]*balanceTask, 0, 200 )
}

func balanceDaemon() {
	for {
		if bCurrent < len( queueBalance ) {
			BalanceProceed()
		}
		utils.Sleep( 10 )
	}
}

func BalanceProceed() {
	bMutex.Lock()
	task := queueBalance[ bCurrent ]
	subject := new(bytes.Buffer)
	html := new(bytes.Buffer)
	text := new(bytes.Buffer)
	pattern := `balance`
//			text := new(bytes.Buffer)
			
	if err := GPagePattern.ExecuteTemplate(subject, pattern + `Subject`, nil ); err != nil {
		task.Error = err
	} 
	// Защита от повторной рассылки
	for i:=0; i<bCurrent; i++ {
		if queueBalance[i].Error == nil && queueBalance[i].UserId == task.UserId {
			task.Error = fmt.Errorf(`It has already been sent`)
			bCurrent++
			bMutex.Lock()
			return
		}
	}
	var user map[string]string

	data := make(map[string]interface{})
	user, task.Error = GDB.OneRow("select * from users where user_id=?", task.UserId ).String()

	if len(user) == 0 {
		task.Error = fmt.Errorf(`The user has no email`)
	} else if utils.StrToInt( user[`verified`] ) < 0 {
		task.Error = fmt.Errorf(`The user in the stop-list`)
	} else {
		data[`Unsubscribe`] = fmt.Sprintf( `%s/unsubscribe?uid=%d-%s`, 
		 	utils.EMAIL_SERVER, task.UserId, strconv.FormatUint( uint64( crc32.ChecksumIEEE([]byte(user[`email`]))), 32 ))
		getBalance( task.UserId, &data )
		if len(data[`List`].(map[int64]*infoBalance)) <= 1 {
			task.Error = fmt.Errorf(`No dcoins`)
			bCurrent++
			bMutex.Unlock()
			return	
		}
		GPagePattern.ExecuteTemplate(html, pattern + `HTML`, data )
	//	GPagePattern.ExecuteTemplate(text, task.Pattern + `Text`, data )
	
		if len( subject.String()) == 0 {
			subject.WriteString(`DCoin notifications`)
		}
		if len(html.String()) > 0 || len(text.String()) > 0 {
			if task.Error == nil {
				bcc := GSettings.CopyTo
				GSettings.CopyTo = ``
				
				if data[`List`].(map[int64]*infoBalance)[72] != nil && data[`List`].(map[int64]*infoBalance)[72].Tdc > 100 {
					task.Error = fmt.Errorf(`Too much Tdc=%f`, data[`List`].(map[int64]*infoBalance)[72].Tdc )
				} else if err := GEmail.SendEmail( html.String(), text.String(), subject.String(),
							[]*Email{&Email{``, user[`email`] }}); err != nil {
					GDB.ExecSql(`update users set verified = -1 where user_id=?`, task.UserId )
					task.Error = err					
				}  else {
	//				log.Println( `Balance Sent:`, user[`email`], userId )
					GDB.ExecSql(`INSERT INTO log ( user_id, email, cmd, params, uptime, ip )
					 VALUES ( ?, ?, ?, ?, datetime('now'), ? )`,
			         task.UserId, user[`email`], utils.ECMD_BALANCE, ``, 1 )
					icur := int64(72)
					if data[`List`].(map[int64]*infoBalance)[icur] == nil || data[`List`].(map[int64]*infoBalance)[icur].Tdc == 0 {
						for icur := range data[`List`].(map[int64]*infoBalance) {
							if icur != 1 {
								break
							}
						}
					}
					if data[`List`].(map[int64]*infoBalance)[icur] != nil {
						task.Error = fmt.Errorf(`Sent Currency=%d Wallet=%f Tdc=%f Promised=%f`, icur,
							data[`List`].(map[int64]*infoBalance)[icur].Wallet,
							data[`List`].(map[int64]*infoBalance)[icur].Tdc,
							data[`List`].(map[int64]*infoBalance)[icur].Promised )
					}
				}
				GSettings.CopyTo = bcc
			}
		} else {
			task.Error = fmt.Errorf(`Wrong HTML and Text patterns`)
		}
	}
	bCurrent++
	bMutex.Unlock()
}

type infoBalance struct {
	Currency   string
	CurrencyId int64
	Wallet     float64
	Tdc        float64
	Promised   float64
	Restricted float64
}

func getBalance( userId int64, data *map[string]interface{} ) error {
	
	list := make(map[int64]*infoBalance)
	if wallet, err := utils.DB.GetBalances(userId); err == nil {
		for _, iwallet := range wallet {
			list[iwallet.CurrencyId] = &infoBalance{ CurrencyId: iwallet.CurrencyId,
			                Wallet: iwallet.Amount }
		}
	} else {
		return err
	}
	if vars, err := utils.DB.GetAllVariables(); err == nil {
		if _, dc, _, err := utils.DB.GetPromisedAmounts( userId, vars.Int64["cash_request_time"]); err == nil {
			for _, idc := range dc {
				if _, ok:= list[idc.CurrencyId]; ok {
					list[idc.CurrencyId].Tdc = idc.Tdc
					list[idc.CurrencyId].Promised = idc.Amount
				} else {
					list[idc.CurrencyId] = &infoBalance{ CurrencyId: idc.CurrencyId,
			                Promised: idc.Amount, Tdc: idc.Tdc }
				}
			}
		} else {
			return err
		}
	} else {
		return err
	}
	c := new(controllers.Controller)
//	c.r = r
//	c.w = w
//	c.sess = sess
//	c.SessRestricted = sessRestricted
	c.SessUserId = userId
	c.DCDB = utils.DB
	if profit,_, err := c.GetPromisedAmountCounter(); err == nil && profit > 0 {
		currency := int64(72)
		if _, ok:= list[currency]; ok {
			list[currency].Restricted = profit - 30
		} else {
			list[currency] = &infoBalance{ CurrencyId: currency,
			                Restricted: profit - 30 }
		}
	}
		
	for i := range list {
		list[i].Currency,_ = utils.DB.Single(`select name from currency where id=?`, list[i].CurrencyId ).String()
	}
	(*data)[`List`] = list
	return nil
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	users := strings.Split( r.PostFormValue(`idusers`), `,` )
	clear := r.PostFormValue(`clearqueueBalance`)
	if len(clear) > 0 {
		bMutex.Lock()
		queueBalance = queueBalance[:0]
		bCurrent = 0
		data[`message`] = `Очередь очищена`
		bMutex.Unlock()
	} else if len(users) > 0 && len(users[0]) > 0 {
		if users[0] == `*` {
			if list, err := GDB.GetAll("select user_id, email from users where verified >= 0", -1); err == nil {
				users = users[:0]
				for _, icur := range list {
					users = append( users, icur[`user_id`])
				}
			}
		}
		bMutex.Lock()
		for _, iduser := range users { 
			queueBalance = append( queueBalance, &balanceTask{ UserId: utils.StrToInt64( iduser ) })		
		}
		bMutex.Unlock()
		http.Redirect(w, r, `/` + GSettings.Admin + `/balance`, http.StatusFound )
	} else {
		data[`message`] = `Не указаны пользователи`
	}
	data[`count`],_ = GDB.Single(`select count(id) from users where verified>=0`).Int64()
	data[`tasks`] = queueBalance[:bCurrent]
	data[`todo`] = len(queueBalance) - bCurrent
	if err := GPageTpl.ExecuteTemplate(out, `balance`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
