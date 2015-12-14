package utils

import (
	"fmt"
	"net"
	"time"
	"sync"
)
// сигнал горутине, которая мониторит таблу chat, что есть новые данные
var ChatNewTx = make(chan bool, 100)
//var ChatJoinConn = make(chan net.Conn)
//var ChatPoolConn []net.Conn
//var ChatDelConn = make(chan net.Conn)

var ChatMutex   = &sync.Mutex{}

type ChatData struct {
	Hashes []byte
	HashesArr [][]byte
}
var ChatDataChan chan *ChatData = make(chan *ChatData, 10)
// исходящие соединения протоколируем тут, используется для подсчета кол-ва
// отправляемых данных в канал ChatDataChan и для исключения создания повторных
// исходящих соединений
var ChatOutConnections map[int64]int = make(map[int64]int)
var ChatInConnections map[int64]int = make(map[int64]int)

// Ждет входящие данные
func ChatInput(conn net.Conn, userId int64) {

	fmt.Println("ChatInput start. wait data from ", conn.RemoteAddr().String(), Time())

	for {

		conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		// тут ждем, пока нам пришлют данные
		fmt.Println("ChatInput for", conn.RemoteAddr().String(), Time())
		binaryData, err := TCPGetSizeAndData(conn, 1048576)
		if err != nil {
			fmt.Println("ChatInput ERROR", err, conn.RemoteAddr().String(), Time())
			safeDeleteFromChatMap(ChatInConnections, userId)
			return
		}
		conn.SetReadDeadline(time.Time{})
		fmt.Printf("binaryData %x\n", binaryData)

		// каждые 30 сек шлется сигнал, что канал еще жив
		if len(binaryData) < 16 {
			fmt.Println(">> Get test data from ", conn.RemoteAddr().String(), Time())
			continue
		}

		var hash []byte
		addsql := ""
		var hashes []map[string]int
		for {
			hash = BytesShift(&binaryData, 16)
			if DB.ConfigIni["db_type"] == "postgresql" {
				addsql += "decode('" + string(BinToHex(hash)) + "', 'hex'),"
			} else {
				addsql += "x'" + string(BinToHex(hash)) + "',"
			}
			hashes = append(hashes, map[string]int{string(hash): 1})
			if len(binaryData) < 16 {
				break
			}
		}

		if len(addsql) == 0 {
			fmt.Println("empty hashes")
			safeDeleteFromChatMap(ChatInConnections, userId)
			return
		}
		addsql = addsql[:len(addsql)-1]
		fmt.Println("addsql", addsql)

		// смотрим в табле chat, чего у нас уже есть
		fmt.Println(`SELECT hash FROM chat WHERE hash IN (`+addsql+`)`)
		rows, err := DB.Query(`SELECT hash FROM chat WHERE hash IN (`+addsql+`)`)
		if err != nil {
			fmt.Println(ErrInfo(err))
			safeDeleteFromChatMap(ChatInConnections, userId)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var hash string
			err = rows.Scan(&hash)
			if err != nil {
				fmt.Println(ErrInfo(err))
				safeDeleteFromChatMap(ChatInConnections, userId)
				return
			}
			// отмечаем 0 то, что у нас уже есть
			for k, v := range hashes {
				if _, ok := v[hash]; ok {
					hashes[k][hash] = 0
				}
			}
		}

		var needTx bool // есть ли что слать
		binHash := ""
		// преобразуем хэши в набор бит, где 0 означет, что такой хэш есть и его слать не надо, а 1 - надо
		for _, hashmap := range hashes {
			for _, result := range hashmap {
				binHash = binHash + IntToStr(result)
				if result == 1 {
					needTx = true
				}
			}
		}

		fmt.Println("binHash", binHash)
		// шлем набор байт, который содержит метки, чего надо качать или "0" - значит ничего качать не будем
		err = WriteSizeAndData([]byte(binHash), conn)
		if err != nil {
			fmt.Println(ErrInfo(err))
			safeDeleteFromChatMap(ChatInConnections, userId)
			return
		}
		if !needTx {
			fmt.Println("continue")
			continue
		}

		// получаем тр-ии, которых у нас нету
		binaryData, err = TCPGetSizeAndData(conn, 10485760)
		if err != nil {
			fmt.Println(ErrInfo(err))
			safeDeleteFromChatMap(ChatInConnections, userId)
			return
		}
		var sendToChan bool
		for {
			length := DecodeLength(&binaryData)
			fmt.Println("length: ", length)
			if int(length) > len(binaryData) {
				fmt.Println("break length > len(binaryData)", length, len(binaryData))
				safeDeleteFromChatMap(ChatInConnections, userId)
				return
			}
			if length > 0 {
				txData := BytesShift(&binaryData, length)
				fmt.Printf("txData %x\n", txData)
				lang := BinToDecBytesShift(&txData, 1)
				room := BinToDecBytesShift(&txData, 4)
				receiver := BinToDecBytesShift(&txData, 4)
				sender := BinToDecBytesShift(&txData, 4)
				status := BinToDecBytesShift(&txData, 1)
				message := BytesShift(&txData, DecodeLength(&txData))
				signTime := BinToDecBytesShift(&txData, 4)
				signature := BinToHex(BytesShift(&txData, DecodeLength(&txData)))

				// проверяем даннные из тр-ий
				err := DB.CheckChatMessage(string(message), sender, receiver, lang, room, status, signTime, signature)
				if err != nil {
					fmt.Println(ErrInfo(err))
					safeDeleteFromChatMap(ChatInConnections, userId)
					return
				}

				data := Int64ToByte(lang)
				data = append(data, Int64ToByte(room)...)
				data = append(data, Int64ToByte(receiver)...)
				data = append(data, Int64ToByte(sender)...)
				data = append(data, Int64ToByte(status)...)
				data = append(data, []byte(message)...)
				data = append(data, Int64ToByte(signTime)...)
				data = append(data, []byte(signature)...)
				hash = Md5(data)
				// заносим в таблу
				err = DB.ExecSql(`INSERT INTO chat (hash, time, lang, room, receiver, sender, status, message, sign_time, signature) VALUES ([hex], ?, ?, ?, ?, ?, ?, ?, ?, [hex])`, hash, Time(), lang, room, receiver, sender, status, message, signTime, signature)
				if err != nil {
					fmt.Println(ErrInfo(err))
					//return
				}
				sendToChan = true

			}
			if length == 0 {
				break
			}
		}

		if sendToChan {
			ChatNewTx <- true
		}
	}
}


// каждый 30 сек шлет данные в канал, чтобы держать его живым
func ChatOutputTesting() {
	for {
		// шлем всем горутинам ChatTxDisseminator, чтобы они разослали по серверам,
		// которые ранее к нам подключились или к которым мы подключались
		//fmt.Println("ChatOutConnections:", ChatOutConnections)
		for i:=0; i < len(ChatOutConnections); i++ {
			ChatDataChan <- nil
		}
		Sleep(30)
	}
}

// ожидает появления свежих записей в чате, затем ждет появления коннектов
// (заносятся из демеона connections и от тех, кто сам подключился к ноде)
func ChatOutput(newTx chan bool) {

	// держим канал в активном состоянии
	go ChatOutputTesting()

	for {
		fmt.Println("ChatOutput wait newTx")
		// просто так тр-ии в chat не появятся, их кто-то должен туда запихать, ждем тут
		<-newTx
		fmt.Println("ChatOutput newTx")

		// смотрим, есть ли в табле неотправленные тр-ии
		rows, err := DB.Query("SELECT hash, lang, room, receiver, sender, status, message, enc_message, sign_time, signature FROM chat WHERE sent = 0 ORDER BY id ASC")
		if err != nil {
			fmt.Println(ErrInfo(err))
		}
		defer rows.Close()
		var hashes []byte
		var hashesArr [][]byte
		for rows.Next() {
			var lang, room, receiver, sender, status, signTime int64
			var message, enc_message string
			var signature, hash []byte
			err = rows.Scan(&hash, &lang, &room, &receiver, &sender, &status, &message, &enc_message, &signTime, &signature)
			if err != nil {
				fmt.Println(ErrInfo(err))
				continue
			}
			if status == 2 {
				message = enc_message
				status = 1
			}
			fmt.Println(`UPDATE chat SET sent = 1 WHERE hex(hash) = ?`, string(BinToHex(hash)))
			err = DB.ExecSql(`UPDATE chat SET sent = 1 WHERE hex(hash) = ?`, string(BinToHex(hash)))
			if err != nil {
				fmt.Println(ErrInfo(err))
				continue
			}
			data := DecToBin(lang, 1)
			data = append(data, DecToBin(room, 4)...)
			data = append(data, DecToBin(receiver, 4)...)
			data = append(data, DecToBin(sender, 4)...)
			data = append(data, DecToBin(status, 1)...)
			data = append(data, EncodeLengthPlusData(message)...)
			data = append(data, DecToBin(signTime, 4)...)
			data = append(data, EncodeLengthPlusData(signature)...)
			//allTx = append(allTx, utils.EncodeLengthPlusData(data))

			hashes = append(hashes, hash...)
			hashesArr = append(hashesArr, data)
		}
		if len(hashes) == 0 {
			fmt.Println("len(hashes) == 0")
			continue
		}

		// шлем всем горутинам ChatTxDisseminator, чтобы они разослали по серверам,
		// которые ранее к нам подключились или к которым мы подключались
		for i:=0; i < len(ChatOutConnections); i++ {
			fmt.Println("ChatData", i, hashes, hashesArr)
			ChatDataChan <- &ChatData{Hashes: hashes, HashesArr: hashesArr}
		}
	}
}

// когда подклюаемся к кому-то или когда кто-то подключается к нам,
// то создается горутина, которая будет ждать, пока появятся свежие
// данные в табле chat, чтобы послать их
func ChatTxDisseminator(conn net.Conn, userId int64) {
	for {
		fmt.Println("wait ChatDataChan send TO->", conn.RemoteAddr().String(), Time())
		data := <-ChatDataChan
		if data == nil {
			fmt.Println("> send test data to ", conn.RemoteAddr().String(), Time())
			err := WriteSizeAndData(EncodeLengthPlusData([]byte{0}), conn)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMap(ChatOutConnections, userId)
				break
			}
			Sleep(1)
			continue
		} else {
			fmt.Println("data", data.Hashes, data.HashesArr, "TO->", conn.RemoteAddr().String(), Time())
			// шлем хэши
			err := WriteSizeAndData(data.Hashes, conn)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMap(ChatOutConnections, userId)
				break
			}
		}
		fmt.Println("WriteSizeAndData ok", conn.RemoteAddr().String(), Time())

		// получаем номера хэшей, тр-ии которых пошлем далее
		hashesBin, err := TCPGetSizeAndData(conn, 10485760)
		if err != nil {
			fmt.Println(ErrInfo(err))
			log.Error("%v", ErrInfo(err))
			safeDeleteFromChatMap(ChatOutConnections, userId)
			break
		}
		fmt.Println("TCPGetSizeAndData ok")

		var TxForSend []byte
		for i := 0; i<len(hashesBin); i++ {
			hashMark := hashesBin[i:i+1]
			if string(hashMark) == "1" {
				TxForSend = append(TxForSend, EncodeLengthPlusData(data.HashesArr[i])...)
			}
		}

		fmt.Printf("TxForSend: %x\n", TxForSend)
		// шлем тр-ии
		if len(TxForSend) > 0 {
			err = WriteSizeAndData(TxForSend, conn)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMap(ChatOutConnections, userId)
				break
			}
		}

		fmt.Println("WriteSizeAndData 2  ok")
		time.Sleep(10 * time.Millisecond)
	}
}

func safeDeleteFromChatMap(delMap map[int64]int, userId int64) {
	ChatMutex.Lock()
	delete(delMap, userId)
	ChatMutex.Unlock()
}