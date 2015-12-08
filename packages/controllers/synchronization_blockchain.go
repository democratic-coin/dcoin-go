package controllers

import (
	"encoding/json"
	"github.com/c-darwin/dcoin-go/packages/consts"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"net/http"
	"os"
	"time"
)

func (c *Controller) SynchronizationBlockchain() (string, error) {

	if c.DCDB == nil || c.DCDB.DB == nil {
		return "", nil
	}
	blockData, err := c.DCDB.GetInfoBlock()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))

		var downloadFile string
		var fileSize int64
		if len(utils.SqliteDbUrl) > 0 {
			downloadFile = *utils.Dir + "/litedb.db"
			resp, err := http.Get(utils.SqliteDbUrl)
			if err != nil {
				return "", err
			}
			fileSize = resp.ContentLength
			resp.Body.Close()
		} else {
			downloadFile = *utils.Dir + "/public/blockchain"
			nodeConfig, err := c.GetNodeConfig()
			blockchain_url := nodeConfig["first_load_blockchain_url"]
			if len(blockchain_url) == 0 {
				blockchain_url = consts.BLOCKCHAIN_URL
			}
			resp, err := http.Get(blockchain_url)
			if err != nil {
				return "", err
			}
			fileSize = resp.ContentLength
		}
		// качается блок
		file, err := os.Open(downloadFile)
		if err != nil {
			return "", err
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			return "", err
		}
		if stat.Size() > 0 {
			log.Debug("stat.Size(): %v", int(stat.Size()))
			return `{"download": "` + utils.Int64ToStr(int64(utils.Round(float64((float64(stat.Size())/float64(fileSize))*100), 0))) + `"}`, nil
		} else {
			return `{"download": "0"}`, nil
		}
	}
	blockId := blockData["block_id"]
	blockTime := blockData["time"]
	if len(blockId) == 0 {
		blockId = "0"
	}
	if len(blockTime) == 0 {
		blockTime = "0"
	}

	wTime := int64(12)
	wTimeReady := int64(2)
	if c.ConfigIni["test_mode"] == "1" {
		wTime = 2 * 365 * 86400
		wTimeReady = 2 * 365 * 86400
	}
	log.Debug("wTime: %v / utils.Time(): %v / blockData[time]: %v", wTime, utils.Time(), utils.StrToInt64(blockData["time"]))
	// если время менее 12 часов от текущего, то выдаем не подвержденные, а просто те, что есть в блокчейне
	if utils.Time()-utils.StrToInt64(blockData["time"]) < 3600*wTime {
		lastBlockData, err := c.DCDB.GetLastBlockData()
		if err != nil {
			return "", err
		}
		log.Debug("lastBlockData[lastBlockTime]: %v", lastBlockData["lastBlockTime"])
		log.Debug("time.Now().Unix(): %v", time.Now().Unix())
		// если уже почти собрали все блоки
		if time.Now().Unix()-lastBlockData["lastBlockTime"] < 3600*wTimeReady {
			blockId = "-1"
			blockTime = "-1"
		}
	}

	connections, err := c.Single(`SELECT count(*) from nodes_connection`).String()
	if err != nil {
		return "", err
	}
	confirmedBlockId, err := c.GetConfirmedBlockId()
	if err != nil {
		return "", err
	}

	currentLoadBlockchain := "nodes"
	if c.NodeConfig["current_load_blockchain"] == "file" {
		currentLoadBlockchain = c.NodeConfig["first_load_blockchain_url"]
	}

	result := map[string]string{"block_id": blockId, "confirmed_block_id": utils.Int64ToStr(confirmedBlockId), "block_time": blockTime, "connections": connections, "current_load_blockchain": currentLoadBlockchain}
	resultJ, _ := json.Marshal(result)

	return string(resultJ), nil
}
