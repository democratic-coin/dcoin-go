package daemons

import (
	"github.com/c-darwin/dcoin-go/packages/utils"
	"os"
	"regexp"
)

func CleaningDb() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "CleaningDb"
	d := new(daemon)
	d.DCDB = DbConnect(GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	if utils.Mobile() {
		d.sleepTime = 1800
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

BEGIN:
	for {
		log.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(GoroutineName) {
			break BEGIN
		}

		curBlockId, err := d.GetBlockId()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// пишем свежие блоки в резервный блокчейн
		endBlockId, err := utils.GetEndBlockId()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			// чтобы не стопориться тут, а дойти до пересборки БД
			endBlockId = 4294967295
		}
		log.Debug("curBlockId: %v / endBlockId: %v", curBlockId, endBlockId)
		if curBlockId-30 > endBlockId {
			file, err := os.OpenFile(*utils.Dir+"/public/blockchain", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			rows, err := d.Query(d.FormatQuery(`
					SELECT id, data
					FROM block_chain
					WHERE id > ? AND id <= ?
					ORDER BY id
					`), endBlockId, curBlockId-30)
			if err != nil {
				file.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			for rows.Next() {
				var id, data string
				err = rows.Scan(&id, &data)
				if err != nil {
					rows.Close()
					file.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				blockData := append(utils.DecToBin(id, 5), utils.EncodeLengthPlusData(data)...)
				sizeAndData := append(utils.DecToBin(len(blockData), 5), blockData...)
				//err := ioutil.WriteFile(*utils.Dir+"/public/blockchain", append(sizeAndData, utils.DecToBin(len(sizeAndData), 5)...), 0644)
				if _, err = file.Write(append(sizeAndData, utils.DecToBin(len(sizeAndData), 5)...)); err != nil {
					rows.Close()
					file.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				if err != nil {
					rows.Close()
					file.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			rows.Close()
			file.Close()
		}

		autoReload, err := d.Single("SELECT auto_reload FROM config").Int64()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		log.Debug("autoReload: %v", autoReload)
		if autoReload < 60 {
			if d.dPrintSleep(utils.ErrInfo("autoReload < 60"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// если main_lock висит более x минут, значит был какой-то сбой
		mainLock, err := d.Single("SELECT lock_time FROM main_lock WHERE script_name NOT IN ('my_lock', 'cleaning_db')").Int64()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		var infoBlockRestart bool
		// если с main_lock всё норм, то возможно, что новые блоки не собираются из-за бана нодов
		if mainLock == 0 || utils.Time()-autoReload < mainLock {
			timeInfoBlock, err := d.Single(`SELECT time FROM info_block`).Int64()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if utils.Time()-timeInfoBlock > autoReload {
				// подождем 5 минут и проверим еще раз
				if d.dSleep(300) {
					break BEGIN
				}
				newTimeInfoBlock, err := d.Single(`SELECT time FROM info_block`).Int64()
				if err != nil {
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				// Если за 5 минут info_block тот же, значит обновление блокчейна не идет
				if newTimeInfoBlock == timeInfoBlock {
					infoBlockRestart = true
				}
			}
		}
		log.Debug("mainLock: %v", mainLock)
		log.Debug("utils.Time(): %v", utils.Time())
		if (mainLock > 0 && utils.Time()-autoReload > mainLock) || infoBlockRestart {
			// на всякий случай пометим, что работаем
			err = d.ExecSql("UPDATE main_lock SET script_name = 'cleaning_db'")
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			err = d.ExecSql("UPDATE config SET pool_tech_works = 1")
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			allTables, err := d.GetAllTables()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			for _, table := range allTables {
				log.Debug("table: %s", table)
				if ok, _ := regexp.MatchString(`^[0-9_]*my_|^e_|install|^config|daemons|payment_systems|community|cf_lang|main_lock`, table); !ok {
					log.Debug("DELETE FROM %s", table)
					err = d.ExecSql("DELETE FROM " + table)
					if err != nil {
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					if table == "cf_currency" {
						if d.ConfigIni["db_type"] == "sqlite" {
							err = d.SetAI("cf_currency", 999)
						} else {
							err = d.SetAI("cf_currency", 1000)
						}
						if err != nil {
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					} else if table == "admin" {
						err = d.ExecSql("INSERT INTO admin (user_id) VALUES (1)")
						if err != nil {
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					} else {
						log.Debug("SET AI %s", table)
						if d.ConfigIni["db_type"] == "sqlite" {
							err = d.SetAI(table, 0)
						} else {
							err = d.SetAI(table, 1)
						}
						if err != nil {
							log.Error("%v", err)
						}
					}
				}
			}
			err = d.ExecSql("DELETE FROM main_lock")
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	log.Debug("break BEGIN %v", GoroutineName)
}
