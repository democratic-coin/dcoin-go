// emailserv
package main

import (
	"net/http"
	"github.com/democratic-coin/dcoin-go/packages/utils"
//	"html/template"
	"bytes"
//	"io"
	"fmt"
	"strings"
	"sync"
)

type sendTask struct {
	UserId  int64
	Pattern string 
	Error   error
}

var (
	stopip map[uint32]byte
	queue     []*sendTask
	qCurrent  int   
	qMutex    sync.Mutex
)

func init() {
	stopip = make(map[uint32]byte)
	queue = make([]*sendTask, 0, 200 )
}

func sendDaemon() {
	for {
		if qCurrent < len( queue ) {
			TaskProceed()
		}
		utils.Sleep( 3 )
	}
}

func TaskProceed() {
	qMutex.Lock()
	task := queue[ qCurrent ]
	
	subject := new(bytes.Buffer)
	html := new(bytes.Buffer)
	text := new(bytes.Buffer)
//			text := new(bytes.Buffer)
			
	if err := GPagePattern.ExecuteTemplate(subject, task.Pattern + `Subject`, nil ); err != nil {
		task.Error = err
	} 
	// Защита от повторной рассылки
	for i:=0; i<qCurrent; i++ {
		if queue[i].Error == nil && queue[i].UserId == task.UserId && queue[i].Pattern == task.Pattern {
			task.Error = fmt.Errorf(`It has already been sent`)
			qCurrent++
			return
		}
	}
	GPagePattern.ExecuteTemplate(html, task.Pattern + `HTML`, nil )
	GPagePattern.ExecuteTemplate(text, task.Pattern + `Text`, nil )
	if len(html.String()) > 0 || len(text.String()) > 0 {
		var user map[string]string
		user, task.Error = GDB.OneRow("select * from users where user_id=?", task.UserId ).String()
		if task.Error == nil {
			bcc := GSettings.CopyTo
			GSettings.CopyTo = ``
			if len(user) == 0 {
				task.Error = fmt.Errorf(`The user has no email`)
			} else if utils.StrToInt( user[`verified`] ) < 0 {
				task.Error = fmt.Errorf(`The user in the stop-list`)
			} else 	if err := GEmail.SendEmail( html.String(), text.String(), subject.String(),
						[]*Email{&Email{``, user[`email`] }}); err != nil {
				GDB.ExecSql(`update users set verified = -1 where user_id=?`, task.UserId )
				task.Error = err					
			}
			GSettings.CopyTo = bcc
		}
	} else {
		task.Error = fmt.Errorf(`Wrong HTML and Text patterns`)
	}
	qCurrent++
	qMutex.Unlock()
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	users := strings.Split( r.PostFormValue(`idusers`), `,` )
	clear := r.PostFormValue(`clearqueue`)
	if len(clear) > 0 {
		qMutex.Lock()
		queue = queue[:0]
		qCurrent = 0
		data[`message`] = `Очередь очищена`
		qMutex.Unlock()
	} else if len(users) > 0 && len(users[0]) > 0 {
		pattern := r.PostFormValue(`pattern`)
		if len(pattern) == 0 {
			data[`message`] = `Не указан шаблон`
		} else {
			if users[0] == `*` {
				if list, err := GDB.GetAll("select user_id from users where verified >= 0", -1); err == nil {
					users = users[:0]
					for _, icur := range list {
						users = append( users, icur[`user_id`])
					}
				}
			}
			qMutex.Lock()
			for _, iduser := range users { 
				queue = append( queue, &sendTask{ UserId: utils.StrToInt64( iduser ), Pattern: pattern })		
			}
			qMutex.Unlock()
			http.Redirect(w, r, `/` + GSettings.Admin + `/send`, http.StatusFound )
		}
	} else {
		data[`message`] = `Не указаны пользователи`
	}
	data[`count`],_ = GDB.Single(`select count(id) from users where verified>=0`).Int64()
	data[`tasks`] = queue[:qCurrent]
	data[`todo`] = len(queue) - qCurrent
	if err := GPageTpl.ExecuteTemplate(out, `send`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
