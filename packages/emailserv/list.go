// emailserv
package main

import (
	"net/http"
	"bytes"
)

func listHandler(w http.ResponseWriter, r *http.Request) {
	
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	GDB.ExecSql(`delete from users where id >=41 AND id<=45`)
	data[`List`],_ = GDB.GetAll(`select user_id, email, verified from users order by user_id`, -1 )
	if err := GPageTpl.ExecuteTemplate(out, `list`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
