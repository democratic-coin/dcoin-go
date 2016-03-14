// emailserv
package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	//	"regexp"
	//	"net/url"
	//	"time"
	"strings"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Settings struct {
	Port      uint32 `json:"port"`
	Path      string `json:"path"`
	ApiId     string `json:"api_id"`
	ApiSecret string `json:"api_secret"`
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
}

type Answer struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

var (
	GSettings Settings
	GDB       *utils.DCDB
	GEmail    *EmailClient
)

func emailHandler(w http.ResponseWriter, r *http.Request) {
	answer := Answer{false, ``}
	cmd := r.URL.Path[1:]
	switch cmd {
	case `setemail`:
		if r.Method != `POST` {
			answer.Error = fmt.Sprintf(`Wrong method %s`, r.Method)
		} else {
			r.ParseForm()
			email := r.FormValue(`email`)
			userid := r.FormValue(`user_id`)
			text := r.FormValue(`text`)
			subject := r.FormValue(`subject`)

			//	re := regexp.MustCompile( `^([a-z0-9_\-]+\.)*[a-z0-9_\-]+@([a-z0-9][a-z0-9\-]*[a-z0-9]\.)+[a-z]{2,4}$` )
			//	if !re.MatchString( email ) {
			if !utils.ValidateEmail(email) {
				answer.Error = fmt.Sprintf(`Incorrect email %s`, email)
				break
			}
			id, _ := utils.DB.Single(`SELECT user_id from users where user_id=?`, userid).String()
			if id != userid {
				answer.Error = fmt.Sprintf(`Incorrect user_id %s`, userid)
				break
			}
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
			var ipval uint32
			log.Println(`IP`, r.RemoteAddr, remoteAddr, ip)
			if ipb := net.ParseIP(remoteAddr).To4(); ipb != nil {
				ipval = uint32(ipb[3]) | (uint32(ipb[2]) << 8) |
					(uint32(ipb[1]) << 16) | (uint32(ipb[0]) << 24)
			}
			err := GDB.ExecSql(`INSERT INTO users ( user_id, email, verified, code, uptime, ip ) 
				VALUES ( ?, ?, 0, 0, datetime('now'), ? )`,
				userid, email, ipval)
			if err != nil {
				answer.Error = fmt.Sprintf(`Command error %v`, err)
				break
			}
			if strings.Index(text, `<`) >= 0 || strings.Index(subject, `<`) >= 0 {
				text = ``
				answer.Error = fmt.Sprintf(`HTML spam`)
			}
			if len(text) > 0 && len(subject) > 0 {
				err = GEmail.SendEmail("<p>"+text+"</p>", text, subject,
					[]*Email{
						&Email{``, email}})
				if err != nil {
					answer.Error = fmt.Sprintf(`Send %v`, err)
				} else {
					log.Println(`Sent test email`, email)
				}

			}
			if len(answer.Error) == 0 {
				answer.Success = true
			} else {
				log.Println(`Error`, answer.Error)
			}
		}

	default:
		answer.Error = fmt.Sprintf(`Unknown command %s`, cmd)
	}
	ret, err := json.Marshal(answer)
	if err != nil {
		ret = []byte(`{"success": false,
"error":"Unknown error"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//	w.WriteHeader(200)
	w.Write(ret)
}

/*func Send() {
	time.Sleep( 5 * time.Second )

	Client := &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   20 * time.Second,
		}
	values := url.Values{}
	values.Set("email", "@mail.ru")
	values.Set("user_id", "1001" )
	values.Set("text", "Test" )
	values.Set("subject", "Test" )
	req, err := http.NewRequest("POST", `http://localhost:8090/setemail`,
	                          strings.NewReader(values.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := Client.Do(req)
	if e != nil {
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var answer Answer
	err = json.Unmarshal( body, &answer )
	if err != nil {
		return
	}
    fmt.Println( answer )
}*/

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	//	os.Chdir(dir)
	logfile, err := os.OpenFile(filepath.Join(dir, "email.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(`Email log`, err)
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
	if len(GSettings.ApiId) == 0 || len(GSettings.ApiSecret) == 0 ||
		len(GSettings.FromEmail) == 0 {
		log.Fatalln(`api_id, api_secret, from_email are not defined`)
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

	if list, err := GDB.GetAllTables(); err == nil && len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE "users" (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	email	TEXT NOT NULL,
	verified	INTEGER NOT NULL,
	code	INTEGER NOT NULL,
	ip	INTEGER NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX userid ON users (user_id)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX ip ON users (ip)`); err != nil {
			log.Fatalln(err)
		}
	}
	GEmail = NewEmailClient(GSettings.ApiId, GSettings.ApiSecret,
		&Email{GSettings.FromName, GSettings.FromEmail})
	log.Println("Start")
	//	go Send()

	http.HandleFunc("/", emailHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	log.Println("Finish")
}
