// emailserv
package main

import (
	"net/http"
	"bytes"
	"fmt"
)

func unbanHandler(w http.ResponseWriter, r *http.Request) {
	
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
		email := r.PostFormValue(`email`)
		if pass != GSettings.Password {
			stopip[ ipval ] += 1
			data[`message`] = `Указан неверный пароль`
		} else if len(email) == 0 {
			data[`message`] = `Не указан email`
		} else {
			stopip[ ipval ] = 0
			if err:= GDB.ExecSql(`update users set verified=0 where email=?`, email ); err != nil {
				data[`message`] = err.Error()
			} else if err:= GDB.ExecSql(`delete from stoplist where email=?`, email ); err != nil {
				data[`message`] = err.Error()
			} else {
				data[`message`] = fmt.Sprintf(`Email %s убран из стоп-листа`, email )
			}
		}
	}
	if err := GPageTpl.ExecuteTemplate(out, `unban`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
