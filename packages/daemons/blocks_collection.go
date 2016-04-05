package daemons

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/democratic-coin/dcoin-go/packages/consts"
	"github.com/democratic-coin/dcoin-go/packages/dcparser"
	"github.com/democratic-coin/dcoin-go/packages/static"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	_ "github.com/lib/pq"
	"os"
)

const GoroutineName = "BlocksCollection"
var d = new(daemon)
var parser = new(dcparser.Parser)
var breaker chan bool
var answer chan string


func BlocksCollection(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()
	breaker = chBreaker
	answer = chAnswer

	d.DCDB = DbConnect(breaker, answer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 300
	} else {
		d.sleepTime = 60
	}
	if !d.CheckInstall(breaker, answer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(breaker, answer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	//var cur bool
BEGIN:
	for {
		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(breaker, answer, GoroutineName) {
			continue
		}
		logger.Debug("0")
		config, err := d.GetNodeConfig()
		if err != nil {
			d.dPrintSleep(err, d.sleepTime)
			continue
		}
		logger.Debug("1")

		// удалим то, что мешает
		if *utils.StartBlockId > 0 {
			del := []string{"queue_tx", "my_notifications", "main_lock"}
			for _, table := range del {
				err := utils.DB.ExecSql(`DELETE FROM `+table)
				fmt.Println(`DELETE FROM `+table)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
			}
		}

		err, restart := d.dbLock()
		if restart {
			logger.Debug("restart true")
			continue
		}
		if err != nil {
			logger.Debug("restart err %v", err)
			d.dPrintSleep(err, d.sleepTime)
			continue
		}
		logger.Debug("2")

		// если это первый запуск во время инсталяции
		currentBlockId, err := d.GetBlockId()
		if err != nil {
			d.unlockPrintSleep(err, d.sleepTime)
			continue
		}

		logger.Info("config", config)
		logger.Info("currentBlockId", currentBlockId)

		// на время тестов
		/*if !cur {
		    currentBlockId = 0
		    cur = true
		}*/

		parser.DCDB = d.DCDB
		parser.GoroutineName = GoroutineName
		if currentBlockId == 0 || *utils.StartBlockId > 0 {
			/*
			   IsNotExistBlockChain := false
			   if _, err := os.Stat(*utils.Dir+"/public/blockchain"); os.IsNotExist(err) {
			       IsNotExistBlockChain = true
			   }*/
			if config["first_load_blockchain"] == "file" /* && IsNotExistBlockChain*/ {

				logger.Info("first_load_blockchain=file")
				nodeConfig, err := d.GetNodeConfig()
				blockchain_url := nodeConfig["first_load_blockchain_url"]
				if len(blockchain_url) == 0 {
					blockchain_url = consts.BLOCKCHAIN_URL
				}
				logger.Debug("blockchain_url: %s", blockchain_url)
				// возможно сервер отдаст блокчейн не с первой попытки
				var blockchainSize int64
				for i := 0; i < 10; i++ {
					logger.Debug("blockchain_url: %s, i: %d", blockchain_url, i)
					blockchainSize, err = utils.DownloadToFile(blockchain_url, *utils.Dir+"/public/blockchain", 3600, breaker, answer, GoroutineName)
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
					}
					if blockchainSize > consts.BLOCKCHAIN_SIZE {
						break
					}
				}
				logger.Debug("blockchain dw ok")
				if err != nil || blockchainSize < consts.BLOCKCHAIN_SIZE {
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
					} else {
						logger.Info(fmt.Sprintf("%v < %v", blockchainSize, consts.BLOCKCHAIN_SIZE))
					}
					if d.unlockPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

				first := true
				/*// блокчейн мог быть загружен ранее. проверим его размер


				  stat, err := file.Stat()
				  if err != nil {
				      if d.unlockPrintSleep(err, d.sleepTime) {	break BEGIN }
				      file.Close()
				      continue BEGIN
				  }
				  if stat.Size() < consts.BLOCKCHAIN_SIZE {
				      d.unlockPrintSleep(fmt.Errorf("%v < %v", stat.Size(), consts.BLOCKCHAIN_SIZE), 1)
				      file.Close()
				      continue BEGIN
				  }*/

				logger.Debug("GO!")
				file, err := os.Open(*utils.Dir + "/public/blockchain")
				if err != nil {
					if d.unlockPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				err = d.ExecSql(`UPDATE config SET current_load_blockchain = 'file'`)
				if err != nil {
					if d.unlockPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

				for {
					// проверим, не нужно ли нам выйти из цикла
					if CheckDaemonsRestart(breaker, answer, GoroutineName) {
						d.unlockPrintSleep(fmt.Errorf("DaemonsRestart"), 0)
						break BEGIN
					}
					b1 := make([]byte, 5)
					file.Read(b1)
					dataSize := utils.BinToDec(b1)
					logger.Debug("dataSize", dataSize)
					if dataSize > 0 {

						data := make([]byte, dataSize)
						file.Read(data)
						//log.Debug("data %x\n", data)
						blockId := utils.BinToDec(data[0:5])
						if *utils.EndBlockId > 0 && blockId == *utils.EndBlockId {
							if d.dPrintSleep(err, 3600) {
								break BEGIN
							}
							file.Close()
							continue BEGIN
						}
						logger.Info("blockId", blockId)
						data2 := data[5:]
						length := utils.DecodeLength(&data2)
						logger.Debug("length", length)
						//log.Debug("data2 %x\n", data2)
						blockBin := utils.BytesShift(&data2, length)
						//log.Debug("blockBin %x\n", blockBin)

						if *utils.StartBlockId == 0 || (*utils.StartBlockId > 0 && blockId > *utils.StartBlockId) {

							// парсинг блока
							parser.BinaryData = blockBin

							if first {
								parser.CurrentVersion = consts.VERSION
								first = false
							}
							err = parser.ParseDataFull()
							if err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								file.Close()
								continue BEGIN
							}
							err = parser.InsertIntoBlockchain()
							if err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								file.Close()
								continue BEGIN
							}

							// отметимся, чтобы не спровоцировать очистку таблиц
							err = parser.UpdMainLock()
							if err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								file.Close()
								continue BEGIN
							}
							if CheckDaemonsRestart(breaker, answer, GoroutineName) {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								file.Close()
								continue BEGIN
							}
						}
						// ненужный тут размер в конце блока данных
						data = make([]byte, 5)
						file.Read(data)
					} else {
						if d.unlockPrintSleep(err, d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					// utils.Sleep(1)
				}
				file.Close()
			} else {

				newBlock, err := static.Asset("static/1block.bin")
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				parser.BinaryData = newBlock
				parser.CurrentVersion = consts.VERSION

				err = parser.ParseDataFull()
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				err = parser.InsertIntoBlockchain()

				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}

			utils.Sleep(1)
			d.dbUnlock()
			continue BEGIN
		}
		d.dbUnlock()

		err = d.ExecSql(`UPDATE config SET current_load_blockchain = 'nodes'`)
		if err != nil {
			d.unlockPrintSleep(err, d.sleepTime)
			continue
		}

		myConfig, err := d.OneRow("SELECT local_gate_ip, static_node_user_id FROM config").String()
		if err != nil {
			d.dPrintSleep(err, d.sleepTime)
			continue
		}

		var hosts []map[string]string
		var nodeHost string
		var dataTypeMaxBlockId, dataTypeBlockBody int64
		if len(myConfig["local_gate_ip"]) > 0 {
			hosts = append(hosts, map[string]string{"host": myConfig["local_gate_ip"], "user_id": myConfig["static_node_user_id"]})
			nodeHost, err = d.Single("SELECT tcp_host FROM miners_data WHERE user_id  =  ?", myConfig["static_node_user_id"]).String()
			if err != nil {
				d.dPrintSleep(err, d.sleepTime)
				continue
			}
			dataTypeMaxBlockId = 9
			dataTypeBlockBody = 8
		} else {
			// получим список нодов, с кем установлено рукопожатие
			hosts, err = d.GetAll("SELECT * FROM nodes_connection", -1)
			if err != nil {
				d.dPrintSleep(err, d.sleepTime)
				continue
			}

			dataTypeMaxBlockId = 10
			dataTypeBlockBody = 7

		}

		logger.Info("%v", hosts)

		if len(hosts) == 0 {
			d.dPrintSleep(err, 1)
			continue
		}

		maxBlockId := int64(1)
		maxBlockIdHost := ""
		var maxBlockIdUserId int64
		// получим максимальный номер блока
		for i := 0; i < len(hosts); i++ {
			if CheckDaemonsRestart(breaker, answer, GoroutineName) {
				break BEGIN
			}
			conn, err := utils.TcpConn(hosts[i]["host"])
			if err != nil {
				if d.dPrintSleep(err, 1) {
					break BEGIN
				}
				continue
			}
			// шлем тип данных
			_, err = conn.Write(utils.DecToBin(dataTypeMaxBlockId, 2))
			if err != nil {
				conn.Close()
				if d.dPrintSleep(err, 1) {
					break BEGIN
				}
				continue
			}
			if len(nodeHost) > 0 { // защищенный режим
				err = utils.WriteSizeAndData([]byte(nodeHost), conn)
				if err != nil {
					conn.Close()
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue
				}
			}
			// в ответ получаем номер блока
			blockIdBin := make([]byte, 4)
			_, err = conn.Read(blockIdBin)
			if err != nil {
				conn.Close()
				if d.dPrintSleep(err, 1) {
					break BEGIN
				}
				continue
			}
			conn.Close()
			id := utils.BinToDec(blockIdBin)
			if id > maxBlockId || i == 0 {
				maxBlockId = id
				maxBlockIdHost = hosts[i]["host"]
				maxBlockIdUserId = utils.StrToInt64(hosts[i]["user_id"])
			}
			if CheckDaemonsRestart(breaker, answer, GoroutineName) {
				utils.Sleep(1)
				break BEGIN
			}
		}

		// получим наш текущий имеющийся номер блока
		// ждем, пока разлочится и лочим сами, чтобы не попасть в тот момент, когда данные из блока уже занесены в БД, а info_block еще не успел обновиться
		err, restart = d.dbLock()
		if restart {
			continue
		}
		if err != nil {
			d.dPrintSleep(err, d.sleepTime)
			continue
		}

		currentBlockId, err = d.GetBlockId()
		if err != nil {
			d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
			continue
		}

		logger.Info("currentBlockId", currentBlockId, "maxBlockId", maxBlockId)
		if maxBlockId <= currentBlockId {
			d.unlockPrintSleep(utils.ErrInfo(errors.New("maxBlockId <= currentBlockId")), d.sleepTime)
			continue
		}

		fmt.Printf("\nnode: %s\n", maxBlockIdHost)

		/////----///////

		if err := collectBlocks(currentBlockId, maxBlockId, dataTypeBlockBody, maxBlockIdUserId, maxBlockIdHost, nodeHost); err != nil {
			continue
		}

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			continue
		}
	}

	logger.Debug("break BEGIN %v", GoroutineName)
}

func collectBlocks(current, max, blockBody, userId int64, host, nodeHost string) error {
	// в цикле собираем блоки, пока не дойдем до максимального
	err := errors.New("Couldn't collect blocks")
	for blockId := current + 1; blockId <= max; blockId++ {
			d.UpdMainLock()
			if CheckDaemonsRestart(breaker, answer, GoroutineName) {
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}
			variables, err := d.GetAllVariables()
			if err != nil {
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}
			// качаем тело блока с хоста maxBlockIdHost
			binaryBlock, err := utils.GetBlockBody(host, blockId, blockBody, nodeHost)

			if len(binaryBlock) == 0 {
				// баним на 1 час хост, который дал нам пустой блок, хотя должен был дать все до максимального
				// для тестов убрал, потом вставить.
				//nodes_ban ($db, $max_block_id_user_id, substr($binary_block, 0, 512)."\n".__FILE__.', '.__LINE__.', '. __FUNCTION__.', '.__CLASS__.', '. __METHOD__);
				//p.NodesBan(maxBlockIdUserId, "len(binaryBlock) == 0")
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}
			binaryBlockFull := binaryBlock
			utils.BytesShift(&binaryBlock, 1) // уберем 1-й байт - тип (блок/тр-я)
			// распарсим заголовок блока
			blockData := utils.ParseBlockHeader(&binaryBlock)
			logger.Info("blockData: %v, blockId: %v", blockData, blockId)

			// если существуют глючная цепочка, тот тут мы её проигнорируем
			badBlocks_, err := d.Single("SELECT bad_blocks FROM config").Bytes()
			if err != nil {
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}
			badBlocks := make(map[int64]string)
			if len(badBlocks_) > 0 {
				err = json.Unmarshal(badBlocks_, &badBlocks)
				if err != nil {
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
				}
			}
			if badBlocks[blockData.BlockId] == string(utils.BinToHex(blockData.Sign)) {
				d.NodesBan(userId, fmt.Sprintf("bad_block = %v => %v", blockData.BlockId, badBlocks[blockData.BlockId]))
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}

			// размер блока не может быть более чем max_block_size
			if current > 1 {
				if int64(len(binaryBlock)) > variables.Int64["max_block_size"] {
					d.NodesBan(userId, fmt.Sprintf(`len(binaryBlock) > variables.Int64["max_block_size"]  %v > %v`, len(binaryBlock), variables.Int64["max_block_size"]))
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
			}

			if blockData.BlockId != blockId {
				d.NodesBan(userId, fmt.Sprintf(`blockData.BlockId != blockId  %v > %v`, blockData.BlockId, blockId))
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}

			// нам нужен хэш предыдущего блока, чтобы проверить подпись
			prevBlockHash := ""
			if blockId > 1 {
				prevBlockHash, err = d.Single("SELECT hash FROM block_chain WHERE id = ?", blockId-1).String()
				if err != nil {
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
				prevBlockHash = string(utils.BinToHex([]byte(prevBlockHash)))
			} else {
				prevBlockHash = "0"
			}
			first := false
			if blockId == 1 {
				first = true
			}
			// нам нужен меркель-рут текущего блока
			mrklRoot, err := utils.GetMrklroot(binaryBlock, variables, first)
			if err != nil {
				d.NodesBan(userId, fmt.Sprintf(`%v`, err))
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}

			// публичный ключ того, кто этот блок сгенерил
			nodePublicKey, err := d.GetNodePublicKey(blockData.UserId)
			if err != nil {
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}

			// SIGN от 128 байта до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
			forSign := fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", blockData.BlockId, prevBlockHash, blockData.Time, blockData.UserId, blockData.Level, mrklRoot)

			// проверяем подпись
			if !first {
				_, err = utils.CheckSign([][]byte{nodePublicKey}, forSign, blockData.Sign, true)
			}

			// качаем предыдущие блоки до тех пор, пока отличается хэш предыдущего.
			// другими словами, пока подпись с prevBlockHash будет неверной, т.е. пока что-то есть в $error
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
				if blockId < 1 {
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
				// нужно привести данные в нашей БД в соответствие с данными у того, у кого качаем более свежий блок
				//func (p *Parser) GetOldBlocks (userId, blockId int64, host string, hostUserId int64, goroutineName, getBlockScriptName, addNodeHost string) error {
				err := parser.GetOldBlocks(blockData.UserId, blockId-1, host, userId, GoroutineName, blockBody, nodeHost)
				if err != nil {
					logger.Error("%v", err)
					d.NodesBan(userId, fmt.Sprintf(`blockId: %v / %v`, blockId, err))
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}

			} else {

				logger.Info("plug found blockId=%v\n", blockId)

				// получим наши транзакции в 1 бинарнике, просто для удобства
				var transactions []byte
				utils.WriteSelectiveLog("SELECT data FROM transactions WHERE verified = 1 AND used = 0")
				rows, err := d.Query("SELECT data FROM transactions WHERE verified = 1 AND used = 0")
				if err != nil {
					utils.WriteSelectiveLog(err)
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
				for rows.Next() {
					var data []byte
					err = rows.Scan(&data)
					utils.WriteSelectiveLog(utils.BinToHex(data))
					if err != nil {
						rows.Close()
						d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
						return err
					}
					transactions = append(transactions, utils.EncodeLengthPlusData(data)...)
				}
				rows.Close()
				if len(transactions) > 0 {
					// отмечаем, что эти тр-ии теперь нужно проверять по новой
					utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
					affect, err := d.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
					if err != nil {
						utils.WriteSelectiveLog(err)
						d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
						return err
					}
					utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
					// откатываем по фронту все свежие тр-ии
					parser.BinaryData = transactions
					err = parser.ParseDataRollbackFront(false)
					if err != nil {
						utils.Sleep(1)
						return err
					}
				}

				err = parser.RollbackTransactionsTestblock(true)
				if err != nil {
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
				err = d.ExecSql("DELETE FROM testblock")
				if err != nil {
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
			}

			// теперь у нас в таблицах всё тоже самое, что у нода, у которого качаем блок
			// и можем этот блок проверить и занести в нашу БД
			parser.BinaryData = binaryBlockFull
			err = parser.ParseDataFull()
			if err == nil {
				err = parser.InsertIntoBlockchain()
				if err != nil {
					d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
					return err
				}
			}
			// начинаем всё с начала уже с другими нодами. Но у нас уже могут быть новые блоки до $block_id, взятые от нода, которого с в итоге мы баним
			if err != nil {
				d.NodesBan(userId, fmt.Sprintf(`blockId: %v / %v`, blockId, err))
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime)
				return err
			}
		}

	return nil
}
