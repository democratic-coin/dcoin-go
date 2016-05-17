// emailserv
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Email struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type EmailClient struct {
	apiId       string
	apiSecret   string
	from        *Email
	timeExpired time.Time
	token       string
	Client      *http.Client
}

type jsonToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint32 `json:"expires_in"`
}

type emailJson struct {
	Html    string   `json:"html"`
	Text    string   `json:"text"`
	Subject string   `json:"subject"`
	From    *Email   `json:"from"`
	To      []*Email `json:"to"`
	Bcc     []*Email `json:"bcc"`
}

const (
	URL_SEND  = "https://api.sendpulse.com/smtp/emails"
	URL_TOKEN = "https://api.sendpulse.com/oauth/access_token"
)

func NewEmailClient(apiId, apiSecret string, from *Email) *EmailClient {
	Client := &EmailClient{
		apiId:     apiId,
		apiSecret: apiSecret,
		from:      from,
		Client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   20 * time.Second,
		},
	}
	return Client
}

func (ec *EmailClient) GetToken() error {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", ec.apiId)
	values.Set("client_secret", ec.apiSecret)

	req, err := http.NewRequest("POST", URL_TOKEN,
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := ec.Client.Do(req)
	if e != nil {
		return e
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var atoken jsonToken
	err = json.Unmarshal(body, &atoken)
	if err != nil {
		return err
	}
	log.Println(`GetToken Success`)
	ec.timeExpired = time.Now().Add(time.Duration(atoken.ExpiresIn) * time.Second)
	ec.token = atoken.TokenType + ` ` + atoken.AccessToken
	return nil
}

func (ec *EmailClient) SendEmail(html, text, subj string, to []*Email) error {
	if time.Now().After(ec.timeExpired) {
		err := ec.GetToken()
		if err != nil {
			return err
		}
	}
	values := url.Values{}

	edata := emailJson{
		Html:    base64.StdEncoding.EncodeToString([]byte(html)),
		Text:    text,
		Subject: subj,
		From:    ec.from,
		To:      to,
	}
	if len( GSettings.CopyTo ) > 0 {
		edata.Bcc = []*Email{ &Email{Email: GSettings.CopyTo}}
	}
	serial, err := json.Marshal(edata)
	if err != nil {
		return err
	}
	values.Set("email", string(serial))
	req, err := http.NewRequest("POST", URL_SEND,
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", ec.token)
	res, e := ec.Client.Do(req)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var ret map[string]bool
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}
	if ret[`result`] {
		return nil
	}
	return fmt.Errorf("%s", body)
}
