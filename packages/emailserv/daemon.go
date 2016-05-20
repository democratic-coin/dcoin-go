// emailserv
package main

import (
	"fmt"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"log"
	"bytes"
	"strconv"
	"hash/crc32"
)

func sendEmail( pattern string, cmd int, userId int64, data *map[string]interface{} ) bool {
	
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
	lang := 0
	if country, err := utils.DB.Single(`select country from miners_data where user_id=?`, userId ).Int64(); err == nil {
		switch country {
			case 10, 14, 19, 67, 80, 112, 119, 125, 180, 214, 224, 230, 235:
				lang = 42
		}	
	}
	(*data)[`Unsubscribe`] = fmt.Sprintf( `%s/unsubscribe?uid=%d-%s`, 
		 	utils.EMAIL_SERVER, userId, strconv.FormatUint( uint64( crc32.ChecksumIEEE([]byte(user[`email`]))), 32 ))
 
	subject := new(bytes.Buffer)
	html := new(bytes.Buffer)
	text := new(bytes.Buffer)
	if lang > 0 {
		GPagePattern.ExecuteTemplate(subject, pattern + `Subject` + utils.IntToStr( lang ), data )
		GPagePattern.ExecuteTemplate(html, pattern + `HTML` + utils.IntToStr( lang ), data )
		GPagePattern.ExecuteTemplate(text, pattern + `Text` + utils.IntToStr( lang ), data )	
	
	}
	if len( subject.String()) == 0 {
		GPagePattern.ExecuteTemplate(subject, pattern + `Subject`, data )
	}
	if len( html.String()) == 0 {
		GPagePattern.ExecuteTemplate(html, pattern + `HTML`, data )
	}
	if len( text.String()) == 0 {
		GPagePattern.ExecuteTemplate(text, pattern + `Text`, data )	
	}
	if len( subject.String()) == 0 {
		subject.WriteString(`DCoin notifications`)
	}
	if len( text.String()) == 0 && len( html.String()) == 0 {
		return result( `Empty pattern ` + pattern )
	}
	if err := GEmail.SendEmail( html.String(), text.String(), subject.String(),
		[]*Email{&Email{``, user[`email`] }}); err != nil {
		GDB.ExecSql(`UPDATE users SET verified=? WHERE user_id=?`, -1, userId )
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
//	sendEmail( `cashreq`, utils.ECMD_CASHREQ, utils.StrToInt64( `0` ), 
//			       &map[string]interface{}{ `Amount`: `2.34`, `Currency`: `USD`, `FromUserId`: `0` })		
	
	for {
		if cash, err := utils.DB.OneRow(`SELECT cash.id, cur.name as currency, from_user_id, to_user_id, currency_id, amount FROM cash_requests as cash
					LEFT JOIN currency as cur ON cur.id=cash.currency_id
		            WHERE cash.id>? order by cash.id`, 
		                 latest[utils.ECMD_CASHREQ] ).String(); err==nil && len(cash) > 0 {
            last := utils.StrToInt64( cash[`id`])							
			if err = GDB.ExecSql(`UPDATE latest SET latest=? WHERE cmd_id=?`, last, utils.ECMD_CASHREQ ); err!=nil {
				log.Println( err )
			}
			sendEmail( `cashreq`, utils.ECMD_CASHREQ, utils.StrToInt64( cash[`to_user_id`] ), 
			       &map[string]interface{}{ `Amount`: cash[`amount`], `Currency`: cash[`currency`], `FromUserId`: cash[`from_user_id`] })		
			latest[utils.ECMD_CASHREQ] = last
		}
		utils.Sleep( 10 )
	}
}
