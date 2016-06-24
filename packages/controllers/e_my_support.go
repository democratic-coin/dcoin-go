package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
/*	"encoding/base64"
	"sort"
	"strings"
	"time"
	"math"*/
)

type eMySupportPage struct {
	Lang             map[string]string
	UserId           int64
}

func (c *Controller) EMySupport() (string, error) {
	var err error

	if c.SessUserId == 0 {
		return `<script language="javascript"> window.location.href = "` + c.EURL + `"</script>If you are not redirected automatically, follow the <a href="` + c.EURL + `">` + c.EURL + `</a>`, nil
	}

	TemplateStr, err := makeTemplate("e_my_support", "eMySupport", &eMySupportPage{
		Lang:             c.Lang,
		UserId:           c.SessUserId,
//		Collapse:         collapse,
	})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
