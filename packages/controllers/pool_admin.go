package controllers

import (
	"errors"
	"github.com/c-darwin/dcoin-go/packages/consts"
	"github.com/c-darwin/dcoin-go/packages/utils"
)

type poolAdminPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Config       map[string]string
	WaitingList  []map[string]string
	UserId       int64
	Lang         map[string]string
	Users        []map[int64]map[string]string
}

func (c *Controller) PoolAdminControl() (string, error) {

	if !c.PoolAdmin {
		return "", utils.ErrInfo(errors.New("access denied"))
	}

	allTable, err := c.GetAllTables()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// удаление юзера с пула
	delId := int64(utils.StrToFloat64(c.Parameters["del_id"]))
	if delId > 0 {

		for _, table := range consts.MyTables {
			if utils.InSliceString(utils.Int64ToStr(delId)+"_"+table, allTable) {
				err = c.ExecSql("DROP TABLE " + utils.Int64ToStr(delId) + "_" + table)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
		}
		err = c.ExecSql("DELETE FROM community WHERE user_id = ?", delId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	if _, ok := c.Parameters["pool_tech_works"]; ok {
		poolTechWorks := int64(utils.StrToFloat64(c.Parameters["pool_tech_works"]))
		poolMaxUsers := int64(utils.StrToFloat64(c.Parameters["pool_max_users"]))
		commission := c.Parameters["commission"]

		//if len(commission) > 0 && !utils.CheckInputData(commission, "commission") {
		//	return "", utils.ErrInfo(errors.New("incorrect commission"))
		//}
		err = c.ExecSql("UPDATE config SET pool_tech_works = ?, pool_max_users = ?, commission = ?", poolTechWorks, poolMaxUsers, commission)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	community, err := c.GetCommunityUsers() // получаем новые данные, т.к. выше было удаление
	var users []map[int64]map[string]string
	for _, uid := range community {
		if uid != c.SessUserId {
			if utils.InSliceString(utils.Int64ToStr(uid)+"_my_table", allTable) {
				data, err := c.OneRow("SELECT miner_id, email FROM " + utils.Int64ToStr(uid) + "_my_table LIMIT 1").String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				users = append(users, map[int64]map[string]string{uid: data})
			}
		}
	}
	log.Debug("users", users)

	// лист ожидания попадания в пул
	waitingList, err := c.GetAll("SELECT * FROM pool_waiting_list", -1)

	config, err := c.GetNodeConfig()
	TemplateStr, err := makeTemplate("pool_admin", "poolAdmin", &poolAdminPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		Config:       config,
		Users:        users,
		UserId:       c.SessUserId,
		WaitingList:  waitingList,
		CountSignArr: c.CountSignArr})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
