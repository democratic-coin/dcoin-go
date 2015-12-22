package daemons

import (
	"github.com/c-darwin/dcoin-go/packages/utils"
	//"log"
	"encoding/json"
	"fmt"
	"github.com/c-darwin/dcoin-go/packages/dcparser"
)

/*
 * Каждые 2 недели собираем инфу о голосах за % и создаем тр-ию, которая
 * попадет в DC сеть только, если мы окажемся генератором блока
 * */
func PctGenerator() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "PctGenerator"
	d := new(daemon)
	d.DCDB = DbConnect(GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	if utils.Mobile() {
		d.sleepTime = 3600
	} else {
		d.sleepTime = 60
	}
	if !d.CheckInstall(DaemonCh, AnswerDaemonCh, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(GoroutineName)
	if d.DCDB == nil {
		return
	}

	err = d.notMinerSetSleepTime(1800)
	if err != nil {
		log.Error("%v", err)
		return
	}

BEGIN:
	for {
		log.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(GoroutineName) {
			break BEGIN
		}

		err, restart := d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		blockId, err := d.GetBlockId()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if blockId == 0 {
			if d.unlockPrintSleep(utils.ErrInfo("blockId == 0"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		_, _, myMinerId, _, _, _, err := d.TestBlock()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		// а майнер ли я ?
		if myMinerId == 0 {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		variables, err := d.GetAllVariables()
		curTime := utils.Time()

		// проверим, прошло ли 2 недели с момента последнего обновления pct
		pctTime, err := d.Single("SELECT max(time) FROM pct").Int64()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if curTime-pctTime > variables.Int64["new_pct_period"] {

			// берем все голоса miner_pct
			pctVotes := make(map[int64]map[string]map[string]int64)
			rows, err := d.Query("SELECT currency_id, pct, count(user_id) as votes FROM votes_miner_pct GROUP BY currency_id, pct ORDER BY currency_id, pct ASC")
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			for rows.Next() {
				var currency_id, votes int64
				var pct string
				err = rows.Scan(&currency_id, &pct, &votes)
				if err != nil {
					rows.Close()
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				log.Info("%v", "newpctcurrency_id", currency_id, "pct", pct, "votes", votes)
				if len(pctVotes[currency_id]) == 0 {
					pctVotes[currency_id] = make(map[string]map[string]int64)
				}
				if len(pctVotes[currency_id]["miner_pct"]) == 0 {
					pctVotes[currency_id]["miner_pct"] = make(map[string]int64)
				}
				pctVotes[currency_id]["miner_pct"][pct] = votes
			}
			rows.Close()

			// берем все голоса user_pct
			rows, err = d.Query("SELECT currency_id, pct, count(user_id) as votes FROM votes_user_pct GROUP BY currency_id, pct ORDER BY currency_id, pct ASC")
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			for rows.Next() {
				var currency_id, votes int64
				var pct string
				err = rows.Scan(&currency_id, &pct, &votes)
				if err != nil {
					rows.Close()
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				log.Info("%v", "currency_id", currency_id, "pct", pct, "votes", votes)
				if len(pctVotes[currency_id]) == 0 {
					pctVotes[currency_id] = make(map[string]map[string]int64)
				}
				if len(pctVotes[currency_id]["user_pct"]) == 0 {
					pctVotes[currency_id]["user_pct"] = make(map[string]int64)
				}
				pctVotes[currency_id]["user_pct"][pct] = votes
			}
			rows.Close()

			newPct := make(map[string]map[string]map[string]string)
			newPct["currency"] = make(map[string]map[string]string)
			var userMaxKey int64
			PctArray := utils.GetPctArray()

			log.Info("%v", "pctVotes", pctVotes)
			for currencyId, data := range pctVotes {

				currencyIdStr := utils.Int64ToStr(currencyId)
				// определяем % для майнеров
				pctArr := utils.MakePctArray(data["miner_pct"])
				log.Info("%v", "pctArrminer_pct", pctArr, currencyId)
				key := utils.GetMaxVote(pctArr, 0, 390, 100)
				log.Info("%v", "key", key)
				if len(newPct["currency"][currencyIdStr]) == 0 {
					newPct["currency"][currencyIdStr] = make(map[string]string)
				}
				newPct["currency"][currencyIdStr]["miner_pct"] = utils.GetPctValue(key)

				// определяем % для юзеров
				pctArr = utils.MakePctArray(data["user_pct"])
				log.Info("%v", "pctArruser_pct", pctArr, currencyId)

				log.Info("%v", "newPct", newPct)
				pctY := utils.ArraySearch(newPct["currency"][currencyIdStr]["miner_pct"], PctArray)
				log.Info("%v", "newPct[currency][currencyIdStr][miner_pct]", newPct["currency"][currencyIdStr]["miner_pct"])
				log.Info("%v", "PctArray", PctArray)
				log.Info("%v", "miner_pct $pct_y=", pctY)
				maxUserPctY := utils.Round(utils.StrToFloat64(pctY)/2, 2)
				userMaxKey = utils.FindUserPct(int(maxUserPctY))
				log.Info("%v", "maxUserPctY", maxUserPctY, "userMaxKey", userMaxKey, "currencyIdStr", currencyIdStr)
				// отрезаем лишнее, т.к. поиск идет ровно до макимального возможного, т.е. до miner_pct/2
				pctArr = utils.DelUserPct(pctArr, userMaxKey)
				log.Info("%v", "pctArr", pctArr)

				key = utils.GetMaxVote(pctArr, 0, userMaxKey, 100)
				log.Info("%v", "data[user_pct]", data["user_pct"])
				log.Info("%v", "pctArr", pctArr)
				log.Info("%v", "userMaxKey", userMaxKey)
				log.Info("%v", "key", key)
				newPct["currency"][currencyIdStr]["user_pct"] = utils.GetPctValue(key)
				log.Info("%v", "user_pct", newPct["currency"][currencyIdStr]["user_pct"])
			}

			newPct_ := new(newPctType)
			newPct_.Currency = make(map[string]map[string]string)
			newPct_.Currency = newPct["currency"]
			newPct_.Referral = make(map[string]int64)
			refLevels := []string{"first", "second", "third"}
			for i := 0; i < len(refLevels); i++ {
				level := refLevels[i]
				var votesReferral []map[int64]int64
				// берем все голоса
				rows, err := d.Query("SELECT " + level + ", count(user_id) as votes FROM votes_referral GROUP BY " + level + " ORDER BY " + level + " ASC ")
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				for rows.Next() {
					var level_, votes int64
					err = rows.Scan(&level_, &votes)
					if err != nil {
						rows.Close()
						if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					votesReferral = append(votesReferral, map[int64]int64{level_: votes})
				}
				rows.Close()
				newPct_.Referral[level] = (utils.GetMaxVote(votesReferral, 0, 30, 10))
			}
			jsonData, err := json.Marshal(newPct_)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			_, myUserId, _, _, _, _, err := d.TestBlock()
			forSign := fmt.Sprintf("%v,%v,%v,%s", utils.TypeInt("NewPct"), curTime, myUserId, jsonData)
			log.Debug("forSign = %v", forSign)
			binSign, err := d.GetBinSign(forSign, myUserId)
			log.Debug("binSign = %x", binSign)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			data := utils.DecToBin(utils.TypeInt("NewPct"), 1)
			data = append(data, utils.DecToBin(curTime, 4)...)
			data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(myUserId))...)
			data = append(data, utils.EncodeLengthPlusData(jsonData)...)
			data = append(data, utils.EncodeLengthPlusData([]byte(binSign))...)

			err = d.InsertReplaceTxInQueue(data)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// и не закрывая main_lock переводим нашу тр-ию в verified=1, откатив все несовместимые тр-ии
			// таким образом у нас будут в блоке только актуальные голоса.
			// а если придет другой блок и станет verified=0, то эта тр-ия просто удалится.

			p := new(dcparser.Parser)
			p.DCDB = d.DCDB
			err = p.TxParser(utils.HexToBin(utils.Md5(data)), data, true)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}
		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	log.Debug("break BEGIN %v", GoroutineName)
}

type newPctType struct {
	Currency map[string]map[string]string `json:"currency"`
	Referral map[string]int64             `json:"referral"`
}
