package controllers

import (
//	"github.com/democratic-coin/dcoin-go/packages/utils"
//	"strings"
	"html"
	"fmt"
)

func (c *Controller) ETicket() (string, error) {

	if c.SessUserId == 0 {
		return c.Lang["sign_up_please"], nil
	}
	c.r.ParseForm()
	subject := html.EscapeString(c.r.FormValue("subject"))
	topic := html.EscapeString(c.r.FormValue("topic"))
	err := c.ExecSql(`insert into e_tickets (user_id, subject, topic, idroot, time, status) 
	                 values(?,?,?,0,datetime('now'), 0 )`, c.SessUserId, subject, topic )
	fmt.Println(`ETicket`, c.SessUserId, err, subject, topic)
	return `1`, nil
}
