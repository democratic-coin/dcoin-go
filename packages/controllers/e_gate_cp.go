package controllers

import (
	"fmt"
)

func (c *Controller) EGateCP() (string, error) {

	c.r.ParseForm()

	fmt.Println(c.r.Form)
	log.Error("EGateCP %v", c.r.Form)

	fmt.Println(c.r.Header.Get("HTTP_HMAC"))
	log.Error("HTTP_HMAC %v", c.r.Header.Get("HTTP_HMAC"))

	fmt.Println(c.r.Header.Get("PHP_AUTH_USER"))
	log.Error("PHP_AUTH_USER %v", c.r.Header.Get("PHP_AUTH_USER"))

	fmt.Println(c.r.Header.Get("PHP_AUTH_PW"))
	log.Error("PHP_AUTH_PW %v", c.r.Header.Get("PHP_AUTH_PW"))

	return ``, nil
}
