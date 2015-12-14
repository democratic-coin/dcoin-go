package controllers

import (
	"encoding/json"
	"errors"
	"github.com/c-darwin/dcoin-go/packages/consts"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"io/ioutil"
)

type nodeConfigPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Config       map[string]string
	WaitingList  []map[string]string
	MyStatus     string
	MyMode       string
	ConfigIni    string
	UserId       int64
	Lang         map[string]string
	EConfig    map[string]string
	Users        []map[int64]map[string]string
	ModeError string
}

func (c *Controller) NodeConfigControl() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	log.Debug("c.Parameters", c.Parameters)
	if _, ok := c.Parameters["save_config"]; ok {
		err := c.ExecSql("UPDATE config SET in_connections_ip_limit = ?, in_connections = ?, out_connections = ?, cf_url = ?, pool_url = ?, pool_admin_user_id = ?, exchange_api_url = ?, auto_reload = ?, http_host = ?, chat_enabled = ?, analytics_disabled = ?, auto_update = ?, auto_update_url = ?", c.Parameters["in_connections_ip_limit"], c.Parameters["in_connections"], c.Parameters["out_connections"], c.Parameters["cf_url"], c.Parameters["pool_url"], c.Parameters["pool_admin_user_id"], c.Parameters["exchange_api_url"], c.Parameters["auto_reload"], c.Parameters["http_host"], c.Parameters["chat_enabled"], c.Parameters["analytics_disabled"], c.Parameters["auto_update"], c.Parameters["auto_update_url"])
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET tcp_listening = ?", c.Parameters["tcp_listening"])
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	if _, ok := c.Parameters["save_e_config"]; ok {
		err := c.ExecSql("DELETE FROM e_config");
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, "enable", c.Parameters["e_enable"]);
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(c.Parameters["e_domain"]) > 0 {
			err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, "domain", c.Parameters["e_domain"]);
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		} else {
			err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, "catalog", c.Parameters["e_catalog"]);
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
		params := []string{"commission", "ps", "pm_s_key", "ik_s_key", "payeer_s_key", "pm_id", "ik_id", "payeer_id", "static_file", "static_file_path", "main_dc_account", "dc_commission", "pm_commission"}
		for _, data := range params {
			err = c.ExecSql(`INSERT INTO e_config (name, value) VALUES (?, ?)`, data, c.Parameters["e_"+data]);
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}

	tcp_listening, err := c.Single(`SELECT tcp_listening FROM ` + c.MyPrefix + `my_table`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	config, err := c.GetNodeConfig()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	config["tcp_listening"] = tcp_listening

	myMode := ""
	modeError := ""
	if _, ok := c.Parameters["switch_pool_mode"]; ok {
		dq := c.GetQuotes()
		log.Debug("c.Community", c.Community)
		if !c.Community { // сингл-мод

			myUserId, err := c.GetMyUserId("")
			commission, err := c.Single("SELECT commission FROM commission WHERE user_id = ?", myUserId).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			// без комиссии не получится генерить блоки и пр., TestBlock() будет выдавать ошибку
			if len(commission) == 0 {
				modeError = "empty commission"
				myMode = "Single"
			} else {
				// переключаемся в пул-мод
				for _, table := range consts.MyTables {

					err = c.ExecSql("ALTER TABLE " + dq + table + dq + " RENAME TO " + dq + utils.Int64ToStr(myUserId) + "_" + table + dq)
					if err != nil {
						return "", utils.ErrInfo(err)
					}
				}
				err = c.ExecSql("INSERT INTO community (user_id) VALUES (?)", myUserId)
				if err != nil {
					return "", utils.ErrInfo(err)
				}


				log.Debug("UPDATE config SET pool_admin_user_id = ?, pool_max_users = 100, commission = ?", myUserId, commission)
				err = c.ExecSql("UPDATE config SET pool_admin_user_id = ?, pool_max_users = 100, commission = ?", myUserId, commission)
				if err != nil {
					return "", utils.ErrInfo(err)
				}

				// восстановим тех, кто ранее был на пуле
				backup_community, err := c.Single("SELECT data FROM backup_community").Bytes()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				if len(backup_community) > 0 {
					var community []int
					err = json.Unmarshal(backup_community, &community)
					if err != nil {
						return "", utils.ErrInfo(err)
					}
					for i := 0; i< len(community); i++ {
						// тут дубль при инсерте, поэтому без обработки ошибок
						c.ExecSql("INSERT INTO community (user_id) VALUES (?)", community[i])
					}
				}
				myMode = "Pool"
			}
		} else {

			// бэкап, чтобы при возврате пул-мода, можно было восстановить
			communityUsers := c.CommunityUsers
			jsonData, _ := json.Marshal(communityUsers)
			backup_community, err := c.Single("SELECT data FROM backup_community").String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if len(backup_community) > 0 {
				err := c.ExecSql("UPDATE backup_community SET data = ?", jsonData)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			} else {
				err = c.ExecSql("INSERT INTO backup_community (data) VALUES (?)", jsonData)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
			myUserId, err := c.GetPoolAdminUserId()
			for _, table := range consts.MyTables {
				err = c.ExecSql("ALTER TABLE " + dq + utils.Int64ToStr(myUserId) + "_" + table + dq + " RENAME TO " + dq + table + dq)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
			err = c.ExecSql("DELETE FROM community")
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			myMode = "Single"
		}
	}

	scriptName, err := c.Single("SELECT script_name FROM main_lock").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	myStatus := "ON"
	if scriptName == "my_lock" {
		myStatus = "OFF"
	}
	if myMode == "" && c.Community {
		myMode = "Pool"
	} else if myMode == "" {
		myMode = "Single"
	}

	configIni, err := ioutil.ReadFile(*utils.Dir + "/config.ini")
	if err != nil {
		return "", utils.ErrInfo(err)
	}




	eConfig, err := c.GetMap(`SELECT * FROM e_config`, "name", "value")
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("node_config", "nodeConfig", &nodeConfigPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		Config:       config,
		UserId:       c.SessUserId,
		MyStatus:     myStatus,
		MyMode:       myMode,
		ConfigIni:    string(configIni),
		EConfig: eConfig,
		ModeError: modeError,
		CountSignArr: c.CountSignArr})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
