// check_blocks
package daemons

import (
	"fmt"
	"time"
	"strings"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"encoding/json"
/*	"errors"
	"github.com/democratic-coin/dcoin-go/packages/consts"
	"github.com/democratic-coin/dcoin-go/packages/dcparser"
	"github.com/democratic-coin/dcoin-go/packages/static"
	_ "github.com/lib/pq"
	"os"*/
)

var (
	checkId int64       // The latest checked block
	checkTime time.Time // The time of the previous comparison
)

func CheckBlocks() {
	defer time.AfterFunc( 30*time.Second, CheckBlocks )
	if utils.DB == nil || utils.DB.DB == nil {
		return
	}
	current, err := utils.DB.GetBlockId()
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
	}
	if current - checkId < 5 && checkTime.Add(15*time.Minute).After( time.Now()){
		return
	}

	fmt.Println(`Checked`, checkId, checkTime, time.Now() )

	q := "SELECT http_host FROM miners_data WHERE miner_id > 0 GROUP BY http_host"
	if configIni["db_type"] == "postgresql" {
		q = "SELECT DISTINCT ON (http_host) http_host FROM miners_data WHERE miner_id > 0"
	}
	hosts, err := utils.DB.GetAll( q, 20 )
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
	}
	for _, item := range hosts {
		host := item[`http_host`]
//		host = `http://localhost:8089/` // !!! только для теста
		if !strings.HasPrefix( host, `http`) {
			continue
		}
		jsonData, err := utils.GetHttpTextAnswer( host + "/ajaxjson?controllerName=CheckHash&block_id="+
		                                          utils.Int64ToStr(current))
		
		if err != nil {
			continue
		}
		var jsonMap map[string]string
		err = json.Unmarshal([]byte(jsonData), &jsonMap)
		if err != nil || jsonMap == nil {
			continue
		}
		fmt.Println(`host`, host, jsonMap )
//		break // !!! Только для теста
	}
	checkTime = time.Now()
	checkId = current
}
