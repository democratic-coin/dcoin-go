package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

type voteForMePage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	UserId       int64
	Lang         map[string]string
	CountSignArr []int
	MyComments   []map[string]string
}

func (c *Controller) VoteForMe() (string, error) {

	// список отравленных нами запросов
	myComments, err := c.GetAll("SELECT * FROM "+c.MyPrefix+"my_comments WHERE comment != 'null' AND type NOT IN ('arbitrator','seller')", -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("vote_for_me", "voteForMe", &voteForMePage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		SignData:     "",
		MyComments:   myComments})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
