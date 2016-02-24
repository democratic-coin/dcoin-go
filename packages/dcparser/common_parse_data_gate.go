package dcparser

import (
	"errors"
	"fmt"
	"github.com/c-darwin/dcoin-go/packages/consts"
	"github.com/c-darwin/dcoin-go/packages/utils"
)

/**
Обработка данных (блоков или транзакций), пришедших с гейта. Только проверка.
*/
func (p *Parser) ParseDataGate(onlyTx bool) error {

	var err error
	p.dataPre()
	p.TxIds = []string{}

	p.Variables, err = p.GetAllVariables()
	if err != nil {
		return utils.ErrInfo(err)
	}

	transactionBinaryData := p.BinaryData
	var transactionBinaryDataFull []byte

	log.Debug("p.dataType: %d", p.dataType)
	// если это транзакции (type>0), а не блок (type==0)
	if p.dataType > 0 {

		// проверим, есть ли такой тип тр-ий
		if len(consts.TxTypes[p.dataType]) == 0 {
			return p.ErrInfo("Incorrect tx type " + utils.IntToStr(p.dataType))
		}

		log.Debug("p.dataType: %d", p.dataType)
		transactionBinaryData = append(utils.DecToBin(int64(p.dataType), 1), transactionBinaryData...)
		transactionBinaryDataFull = transactionBinaryData

		// нет ли хэша этой тр-ии у нас в БД?
		err = p.CheckLogTx(transactionBinaryDataFull)
		if err != nil {
			return p.ErrInfo(err)
		}

		p.TxHash = utils.Md5(transactionBinaryData)

		// преобразуем бинарные данные транзакции в массив
		log.Debug("transactionBinaryData: %x", transactionBinaryData)
		p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("p.TxSlice", p.TxSlice)
		if len(p.TxSlice) < 3 {
			return p.ErrInfo(errors.New("len(p.TxSlice) < 3"))
		}

		// время транзакции может быть немного больше, чем время на ноде.
		// у нода может быть просто не настроено время.
		// время транзакции используется только для борьбы с атаками вчерашними транзакциями.
		// А т.к. мы храним хэши в log_transaction за 36 часов, то боятся нечего.
		curTime := utils.Time()
		if utils.BytesToInt64(p.TxSlice[2])-consts.MAX_TX_FORW > curTime || utils.BytesToInt64(p.TxSlice[2]) < curTime-consts.MAX_TX_BACK {
			return p.ErrInfo(errors.New("incorrect tx time"))
		}
		// $this->transaction_array[3] могут подсунуть пустой
		if !utils.CheckInputData(p.TxSlice[3], "bigint") {
			return p.ErrInfo(errors.New("incorrect user id"))
		}
	}

	// если это блок
	if p.dataType == 0 {

		txCounter := make(map[int64]int64)

		// если есть $only_tx=true, то значит идет восстановление уже проверенного блока и заголовок не требуется
		if !onlyTx {
			err = p.ParseBlock()
			if err != nil {
				return p.ErrInfo(err)
			}

			// проверим данные, указанные в заголовке блока
			err = p.CheckBlockHeader()
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		log.Debug("onlyTx", onlyTx)

		// если в ходе проверки тр-ий возникает ошибка, то вызываем откатчик всех занесенных тр-ий. Эта переменная для него
		p.fullTxBinaryData = p.BinaryData
		var txForRollbackTo []byte
		if len(p.BinaryData) > 0 {
			for {
				transactionSize := utils.DecodeLength(&p.BinaryData)
				if len(p.BinaryData) == 0 {
					return utils.ErrInfo(fmt.Errorf("empty BinaryData"))
				}

				// отчекрыжим одну транзакцию от списка транзакций
				transactionBinaryData := utils.BytesShift(&p.BinaryData, transactionSize)
				transactionBinaryDataFull = transactionBinaryData

				// добавляем взятую тр-ию в набор тр-ий для RollbackTo, в котором пойдем в обратном порядке
				txForRollbackTo = append(txForRollbackTo, utils.EncodeLengthPlusData(transactionBinaryData)...)

				// нет ли хэша этой тр-ии у нас в БД?
				err = p.CheckLogTx(transactionBinaryDataFull)
				if err != nil {
					p.RollbackTo(txForRollbackTo, true, false)
					return p.ErrInfo(err)
				}

				p.TxHash = utils.Md5(transactionBinaryData)
				p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
				log.Debug("p.TxSlice %s", p.TxSlice)
				if err != nil {
					p.RollbackTo(txForRollbackTo, true, false)
					return p.ErrInfo(err)
				}

				var userId int64
				// txSlice[3] могут подсунуть пустой
				if len(p.TxSlice) > 3 {
					if !utils.CheckInputData(p.TxSlice[3], "int64") {
						return utils.ErrInfo(fmt.Errorf("empty user_id"))
					} else {
						userId = utils.BytesToInt64(p.TxSlice[3])
					}
				} else {
					return utils.ErrInfo(fmt.Errorf("empty user_id"))
				}

				// считаем по каждому юзеру, сколько в блоке от него транзакций
				txCounter[userId]++

				// чтобы 1 юзер не смог прислать дос-блок размером в 10гб, который заполнит своими же транзакциями
				if txCounter[userId] > p.Variables.Int64["max_block_user_transactions"] {
					p.RollbackTo(txForRollbackTo, true, false)
					return utils.ErrInfo(fmt.Errorf("max_block_user_transactions"))
				}

				// проверим, есть ли такой тип тр-ий
				_, ok := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				if !ok {
					return utils.ErrInfo(fmt.Errorf("nonexistent type"))
				}

				p.TxMap = map[string][]byte{}

				// для статы
				p.TxIds = append(p.TxIds, string(p.TxSlice[1]))

				MethodName := consts.TxTypes[utils.BytesToInt(p.TxSlice[1])]
				log.Debug("MethodName", MethodName+"Init")
				err_ := utils.CallMethod(p, MethodName+"Init")
				if _, ok := err_.(error); ok {
					log.Debug("error: %v", err)
					p.RollbackTo(txForRollbackTo, true, true)
					return utils.ErrInfo(err_.(error))
				}

				log.Debug("MethodName", MethodName+"Front")
				err_ = utils.CallMethod(p, MethodName+"Front")
				if _, ok := err_.(error); ok {
					log.Debug("error: %v", err)
					p.RollbackTo(txForRollbackTo, true, true)
					return utils.ErrInfo(err_.(error))
				}

				// пишем хэш тр-ии в лог
				err = p.InsertInLogTx(transactionBinaryDataFull, utils.BytesToInt64(p.TxMap["time"]))
				if err != nil {
					return utils.ErrInfo(err)
				}

				if len(p.BinaryData) == 0 {
					break
				}
			}
		}
	} else {

		// Оперативные транзакции
		MethodName := consts.TxTypes[p.dataType]
		log.Debug("MethodName", MethodName+"Init")
		err_ := utils.CallMethod(p, MethodName+"Init")
		if _, ok := err_.(error); ok {
			return utils.ErrInfo(err_.(error))
		}

		log.Debug("MethodName", MethodName+"Front")
		err_ = utils.CallMethod(p, MethodName+"Front")
		if _, ok := err_.(error); ok {
			return utils.ErrInfo(err_.(error))
		}

		// пишем хэш тр-ии в лог
		err = p.InsertInLogTx(transactionBinaryDataFull, utils.BytesToInt64(p.TxMap["time"]))
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return nil
}
