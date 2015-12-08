package controllers

import (
	"errors"
	"github.com/c-darwin/dcoin-go/packages/utils"
)

func (c *Controller) SaveHost() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	http_host := c.r.FormValue("http_host")
	if len(http_host) > 0 && http_host[len(http_host)-1:] != "/" {
		http_host += "/"
	}
	tcp_host := c.r.FormValue("tcp_host")

	if !utils.CheckInputData(http_host, "http_host") {
		return `{"error":"1"}`, nil
	}
	if !utils.CheckInputData(tcp_host, "tcp_host") {
		return `{"error":"1"}`, nil
	}

	err := c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET http_host = ?, tcp_host = ?", http_host, tcp_host)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return `{"error":"0"}`, nil
}
