// emailserv
package main

import (
	"fmt"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"log"
)

func sendEmail( text string, cmd int, userId int64 ) bool {
	
	result := func( msg string ) bool {
		log.Println( fmt.Sprintf( `Daemon Error: user_id=%d %s`, userId, msg ))
		return false
	}
	user,err := GDB.OneRow(`SELECT * FROM users WHERE user_id=?`, userId ).String()
	if err != nil {
		return result( err.Error() )
	}
	if len(user) == 0 {
		return result( `No email` )
	}
	if utils.StrToInt(user[`verified`]) < 0 {
		return result( `Stop list`)
	}
	subject := `DCoin notifications`
	if err := GEmail.SendEmail("<p>"+text+"</p>", text, subject,
		[]*Email{&Email{``, user[`email`] }}); err != nil {
		if err = GDB.ExecSql(`UPDATE users SET verified=? WHERE id=?`, -1, userId ); err!=nil {
			return result( err.Error() )
		}
		return result( fmt.Sprintf(`SendPulse %s`, err.Error()))
	}
	log.Println( `Daemon Sent:`, cmd, user[`email`], userId )
	if err := GDB.ExecSql(`INSERT INTO log ( user_id, email, cmd, params, uptime, ip )
				 VALUES ( ?, ?, ?, ?, datetime('now'), ? )`,
		         userId, user[`email`], cmd, ``, 1 ); err != nil {
					return result( err.Error() )
			}
	return true
}

func daemon() {
	latest := make(map[int]int64)
	if curlatest, err := GDB.GetAll(`SELECT * FROM latest`, -1 ); err == nil {
		for _, curi := range curlatest {
			latest[ utils.StrToInt(curi[`cmd_id`])] = utils.StrToInt64(curi[`latest`])
		}
	} else {
		log.Fatalln( err )
	}
	for {
		if cash, err := utils.DB.OneRow(`SELECT cash.id, cur.name as currency, from_user_id, to_user_id, currency_id, amount FROM cash_requests as cash
					LEFT JOIN currency as cur ON cur.id=cash.currency_id
		            WHERE cash.id>? order by cash.id`, 
		                 latest[utils.ECMD_CASHREQ] ).String(); err==nil && len(cash) > 0 {
            last := utils.StrToInt64( cash[`id`])							
			if err = GDB.ExecSql(`UPDATE latest SET latest=? WHERE cmd_id=?`, last, utils.ECMD_CASHREQ ); err!=nil {
				log.Println( err )
			}
			text := fmt.Sprintf(`You"ve got the request for %s %s. It has to be repaid within the next 48 hours.`,
			           cash[`amount`], cash[`currency`])
			sendEmail( text, utils.ECMD_CASHREQ, utils.StrToInt64( cash[`to_user_id`] ))		
			latest[utils.ECMD_CASHREQ] = last
		}
		utils.Sleep( 10 )
	}
}
