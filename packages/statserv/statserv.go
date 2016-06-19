// statserv
package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"io/ioutil"
	"log"
//	"net"
	"net/http"
	"os"
	"path/filepath"
	//	"regexp"
	//	"net/url"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Settings struct {
	Port      uint32 `json:"port"`
	Path      string `json:"path"`
}

var (
	GSettings Settings
	GDB       *utils.DCDB
)

func statHandler(w http.ResponseWriter, r *http.Request) {
	answer := utils.Answer{false, ``}

	result := func(msg string) {

/*		answer.Error = msg
		if !answer.Success {
			if len(jsonEmail.Email) == 0 {
				jsonEmail.Email = r.FormValue(`email`)
			}
			log.Println(remoteAddr, `Error:`, jsonEmail.Cmd, answer.Error, jsonEmail.Email, jsonEmail.UserId)
		} else {
			log.Println(remoteAddr, `Sent:`, jsonEmail.Cmd, jsonEmail.Email, jsonEmail.UserId)
		}
*/
		ret, err := json.Marshal(answer)
		if err != nil {
			ret = []byte(`{"success": false,
"error":"Unknown error"}`)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		//	w.WriteHeader(200)
		w.Write(ret)
	}
/*	iplog, err := GDB.Single(`select count(id) from log where ip=? AND date( uptime, '+1 hour' ) > datetime('now')`, 
	                     ipval ).Int64()
	if err!=nil {
		log.Println("SQL Error", err )
	} else if iplog > 10 {
		result(`Anti-spam`)
		return
	}
	
	r.ParseForm()

	if len(r.URL.Path[1:]) > 0 || r.Method != `POST` {
		result(`Wrong method or path`)
		return
	}

*/
	answer.Success = true

	result(``)
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	//	os.Chdir(dir)
	logfile, err := os.OpenFile(filepath.Join(dir, "stat.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(`Stat log`, err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	params, err := ioutil.ReadFile(filepath.Join(dir, `settings.json`))
	if err != nil {
		log.Fatalln(dir, `Settings.json`, err)
	}
	if err = json.Unmarshal(params, &GSettings); err != nil {
		log.Fatalln(`Unmarshall`, err)
	}
	if err = os.Chdir(GSettings.Path); err != nil {
		log.Fatalln(`Chdir`, err)
	}
	if GDB, err = utils.NewDbConnect(map[string]string{
		"db_name": "", "db_password": ``, `db_port`: ``,
		`db_user`: ``, `db_host`: ``, `db_type`: `sqlite`}); err != nil {
		log.Fatalln(`Connect`, err)
	}

	*utils.Dir = GSettings.Path
	configIni := make(map[string]string)
	configIni_, err := config.NewConfig("ini", `config.ini`)
	if err != nil {
		log.Fatalln(`Config`, err)
	} else {
		configIni, err = configIni_.GetSection("default")
	}
	if utils.DB, err = utils.NewDbConnect(configIni); err != nil {
		log.Fatalln(`Utils connect`, err)
	}
	
	var list []string
	if list, err = GDB.GetAllTables(); err == nil && len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE balance (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	data    TEXT NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX userid ON balance (user_id,uptime)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE TABLE req_balance (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	ip	INTEGER NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX req_userid ON req_balance (user_id)`); err != nil {
			log.Fatalln(err)
		}
	}
	os.Chdir(dir)	
	go daemon()

	log.Println("Start")
	

	http.HandleFunc( `/`, statHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	log.Println("Finish")
}
