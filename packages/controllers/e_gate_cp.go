package controllers

import (
	"fmt"
)

func (c *Controller) EGateCP() (string, error) {

	c.r.ParseForm()

	fmt.Println(c.r.Form)
	log.Error("EGateCP %v", c.r.Form)

	return ``, nil
}
