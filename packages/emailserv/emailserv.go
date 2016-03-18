// emailserv
package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
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
	"strings"
	"time"
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

var (
	GSettings Settings
	GDB       *utils.DCDB
	GEmail    *EmailClient
)

func emailHandler(w http.ResponseWriter, r *http.Request) {

	answer := utils.Answer{false, ``}

	result := func(msg string) {

		answer.Error = msg
		ret, err := json.Marshal(answer)
		if err != nil {
			ret = []byte(`{"success": false,
"error":"Unknown error"}`)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		//	w.WriteHeader(200)
		w.Write(ret)
	}

	if len(r.URL.Path[1:]) > 0 || r.Method != `POST` {
		result(`Wrong method or path`)
		return
	}
	var (
		jsonEmail utils.JsonEmail
		err       error
		publicKey string
	)
	r.ParseForm()
	data := r.FormValue(`data`)
	sign := r.FormValue(`sign`)
	if err = json.Unmarshal([]byte(data), &jsonEmail); err != nil ||
		jsonEmail.UserId == 0 || jsonEmail.Cmd == 0 {
		result(`Incorrect data`)
		return
	}
	//	re := regexp.MustCompile( `^([a-z0-9_\-]+\.)*[a-z0-9_\-]+@([a-z0-9][a-z0-9\-]*[a-z0-9]\.)+[a-z]{2,4}$` )
	//	if !re.MatchString( email ) {
	if !utils.ValidateEmail(jsonEmail.Email) {
		result(`Incorrect email`)
		return
	}

	if publicKey, err = utils.DB.GetUserPublicKey(jsonEmail.UserId); err != nil {
		pubVal := r.FormValue(`public`)
		if jsonEmail.Cmd == utils.ECMD_TEST && len(pubVal) > 0 {
			public, _ := base64.StdEncoding.DecodeString(pubVal)
			publicKey = string(public)
		} else {
			result(`Incorrect user_id or public_key`)
			return
		}
	}
	fmt.Println(jsonEmail)
	signature, _ := base64.StdEncoding.DecodeString(sign)
	var re interface{}
	if re, err = x509.ParsePKIXPublicKey([]byte(publicKey)); err != nil {
		result(err.Error())
		return
	}
	if err = rsa.VerifyPKCS1v15(re.(*rsa.PublicKey), crypto.SHA1, utils.HashSha1(data),
		signature); err != nil {
		result(err.Error())
		return
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
	if ipb := net.ParseIP(remoteAddr).To4(); ipb != nil {
		ipval = uint32(ipb[3]) | (uint32(ipb[2]) << 8) |
			(uint32(ipb[1]) << 16) | (uint32(ipb[0]) << 24)
	}
	fmt.Println(ipval)
	/*			err := GDB.ExecSql(`INSERT INTO users ( user_id, email, verified, code, uptime, ip )
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
	*/
	answer.Success = true

	result(``)
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

func Send() {
	time.Sleep(5 * time.Second)
	fmt.Println("Result", utils.SendEmail(`test@mail.ru`, 3, utils.ECMD_TEST, nil /*&map[string]string{}*/))
}

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
	go Send()

	http.HandleFunc("/", emailHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	log.Println("Finish")
}
