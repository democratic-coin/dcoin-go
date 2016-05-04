// emailserv
package main

import (
	"net/http"
	"bytes"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	if err := GPageTpl.ExecuteTemplate(out, `login`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
