package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
//	"strings"
	"html"
//	"fmt"
)

func (c *Controller) ETicket() (string, error) {

	if c.SessUserId == 0 {
		return c.Lang["sign_up_please"], nil
	}
	c.r.ParseForm()
	subject := html.EscapeString(c.r.FormValue("subject"))
	topic := html.EscapeString(c.r.FormValue("topic"))
	idroot := utils.StrToInt64(c.r.FormValue("idroot"))
	
	err := c.ExecSql(`insert into e_tickets (user_id, subject, topic, idroot, time, status, uptime) 
	                 values(?,?,?,?,datetime('now'), 1,datetime('now'))`, c.SessUserId, subject, topic, idroot )
	if err == nil && idroot>0 {
		c.ExecSql(`update e_tickets set uptime=datetime('now') where id=?`, idroot )
	}			
	return `1`, nil
}
