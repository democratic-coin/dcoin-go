// statserv
package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"io/ioutil"
	"log"
//	"strings"
//	"net"
	"net/http"
	"os"
	"path/filepath"
	"math/rand"
	//	"regexp"
	//	"net/url"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Settings struct {
	IP   string `json:"ip"`
	Port uint32 `json:"port"`
	Path string `json:"path"`
	Urls []string `json:"urls"`  
}

var (
	GSettings Settings
	GDB       *utils.DCDB
)
/*
func getIP(r *http.Request) (uint32, string) {
	var ipval uint32

	remoteAddr := r.RemoteAddr
	var ip string
	if ip = r.Header.Get(XRealIP); len(ip) > 6 {
		remoteAddr = ip
	} else if ip = r.Header.Get(XForwardedFor); len(ip) > 6 {
		remoteAddr = ip
	}
	if strings.Contains(remoteAddr, ":") {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}
	if ipb := net.ParseIP(remoteAddr).To4(); ipb != nil {
		ipval = uint32(ipb[3]) | (uint32(ipb[2]) << 8) |
			(uint32(ipb[1]) << 16) | (uint32(ipb[0]) << 24)
	}
	return ipval,remoteAddr
}
*/

func IndexGetPool(w http.ResponseWriter, r *http.Request) {
	var answer string
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")		

	if utils.DB != nil && utils.DB.DB != nil && len(GSettings.Urls) == 0 {

		var err error
		var poolHttpHost string
		var getUserId int64
		
		publicKey := r.FormValue("public_key")
		if len( publicKey ) > 0 {
			getUserId, err = utils.DB.Single("SELECT user_id FROM users WHERE hex(public_key_0) = ?", publicKey).Int64()
			if err != nil {
				log.Println("%v", err)
			}
		} else {
			getUserId = utils.StrToInt64(r.FormValue("user_id"))
		}
		if getUserId == 0 {
			variables, err := utils.DB.GetAllVariables()
			poolHttpHost, err = utils.DB.Single(`SELECT http_host FROM miners_data WHERE i_am_pool = 1 AND pool_count_users < ?`, variables.Int64["max_pool_users"]).String()
			if err != nil {
				log.Println("%v", err)
			}
		} else {
			poolHttpHost, err = utils.DB.Single("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host end as http_host FROM miners_data as m WHERE m.user_id = ?", getUserId).String()
			if err != nil {
				log.Println("%v", err)
			}
		}
		answer = `{"pool":"`+poolHttpHost+`"}`
		if len( publicKey ) > 0 {
			answer = `{"pool":"`+poolHttpHost+`", "user_id":`+utils.Int64ToStr(getUserId)+`}`
		}
	} else if len(GSettings.Urls) > 0 {
		var ind int
		if len(GSettings.Urls) > 1 {
			rand.Seed(447)
			ind = rand.Intn( len(GSettings.Urls))
		}
		answer = `{"pool":"`+GSettings.Urls[ind]+`"}`
	}
	if _, err := w.Write([]byte(answer)); err != nil {
		log.Println("%v", err)
	}
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	//	os.Chdir(dir)
	logfile, err := os.OpenFile(filepath.Join(dir, "getpool.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(`Getpool log`, err)
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
/*	if GDB, err = utils.NewDbConnect(map[string]string{
		"db_name": "", "db_password": ``, `db_port`: ``,
		`db_user`: ``, `db_host`: ``, `db_type`: `sqlite`}); err != nil {
		log.Fatalln(`Connect`, err)
	}*/

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

	os.Chdir(dir)

	log.Println("Start")

	http.HandleFunc(`/`, IndexGetPool)
	http.ListenAndServe(fmt.Sprintf("%s:%d", GSettings.IP, GSettings.Port), nil)
	log.Println("Finish")
}
