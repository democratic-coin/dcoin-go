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
	task := queue[ qCurrent ]
	
	subject := new(bytes.Buffer)
	html := new(bytes.Buffer)
	text := new(bytes.Buffer)
//			text := new(bytes.Buffer)
			
	if err := GPageTpl.ExecuteTemplate(subject, task.Pattern + `Subject`, nil ); err != nil {
		task.Error = err
	} else if err := GPageTpl.ExecuteTemplate(html, task.Pattern + `HTML`, nil ); err != nil {
		task.Error = err
	} else if err := GPageTpl.ExecuteTemplate(text, task.Pattern + `Text`, nil ); err != nil {
		task.Error = err
	} else {
		var user map[string]string
		user, task.Error = GDB.OneRow("select * from users where user_id=?", task.UserId ).String()
		if task.Error == nil {
			if len(user) == 0 {
				task.Error = fmt.Errorf(`The user has no email`)
			} else if utils.StrToInt( user[`verified`] ) < 0 {
				task.Error = fmt.Errorf(`The user in the stop-list`)
			} else if err = GEmail.SendEmail( html.String(), text.String(), subject.String(),
						[]*Email{&Email{``, user[`email`] }}); err != nil {
				GDB.ExecSql(`update users set verified = -1 where user_id=?`, task.UserId )
				task.Error = err					
			}
		}
	}
	qCurrent++
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	
	ipval,_ := getIP( r )

	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	pass := r.PostFormValue(`password`)
	if len(pass) > 0 {
		if stopip[ ipval ] > 5 {
			w.Write( []byte(`Blocked`))
			return
		}
		users := strings.Split( r.PostFormValue(`idusers`), `,` )
		pattern := r.PostFormValue(`pattern`)
		if pass != GSettings.Password {
			stopip[ ipval ] += 1
			data[`message`] = `Указан неверный пароль`
		} else if len(users) == 0 {
			data[`message`] = `Не указаны пользователи`
		} else if len(pattern) == 0 {
			data[`message`] = `Не указаны шаблон`
		} else {
			stopip[ ipval ] = 0
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
		}
	}
	data[`count`],_ = GDB.Single(`select count(id) from users where verified>=0`).Int64()
	data[`tasks`] = queue[:qCurrent]
	if err := GPageTpl.ExecuteTemplate(out, `send`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
