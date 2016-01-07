package daemons

import (
	"errors"
	"flag"
	"github.com/astaxie/beego/config"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"github.com/op/go-logging"
	"os"
)

var (
	log                           = logging.MustGetLogger("daemons")
	/*DaemonCh        chan bool     = make(chan bool, 100)
	AnswerDaemonCh  chan string   = make(chan string, 100)*/
	MonitorDaemonCh chan []string = make(chan []string, 100)
	configIni       map[string]string
)

type daemon struct {
	*utils.DCDB
	goRoutineName  string
	/*DaemonCh       chan bool
	AnswerDaemonCh chan string*/
	chBreaker chan bool
	chAnswer chan string
	sleepTime      int
}

func (d *daemon) dbLock() (error, bool) {
	return d.DbLock(d.chBreaker, d.chAnswer, d.goRoutineName)
}

func (d *daemon) dbUnlock() error {
	log.Debug("dbUnlock %v", utils.Caller(1))
	return d.DbUnlock(d.goRoutineName)
}

func (d *daemon) dSleep(sleep int) bool {
	for i := 0; i < sleep; i++ {
		if CheckDaemonsRestart(d.chBreaker, d.chAnswer, d.goRoutineName) {
			return true
		}
		utils.Sleep(1)
	}
	return false
}

func (d *daemon) dPrintSleep(err_ interface{}, sleep int) bool {
	var err error
	switch err_.(type) {
	case string:
		err = errors.New(err_.(string))
	case error:
		err = err_.(error)
	}
	log.Error("%v (%v)", err, utils.GetParent())
	if d.dSleep(sleep) {
		return true
	}
	return false
}

func (d *daemon) unlockPrintSleep(err error, sleep int) bool {
	if err != nil {
		log.Error("%v", err)
	}
	err = d.DbUnlock(d.goRoutineName)
	if err != nil {
		log.Error("%v", err)
	}
	for i := 0; i < sleep; i++ {
		if CheckDaemonsRestart(d.chBreaker, d.chAnswer, d.goRoutineName) {
			return true
		}
		utils.Sleep(1)
	}
	return false
}

func (d *daemon) unlockPrintSleepInfo(err error, sleep int) bool {
	if err != nil {
		log.Debug("%v", err)
	}
	err = d.DbUnlock(d.goRoutineName)
	if err != nil {
		log.Error("%v", err)
	}

	for i := 0; i < sleep; i++ {
		if CheckDaemonsRestart(d.chBreaker, d.chAnswer, d.goRoutineName) {
			return true
		}
		utils.Sleep(1)
	}
	return false
}

func (d *daemon) notMinerSetSleepTime(sleep int) error {
	community, err := d.GetCommunityUsers()
	if err != nil {
		return err
	}
	if len(community) == 0 {
		userId, err := d.GetMyUserId("")
		if err != nil {
			return err
		}
		minerId, err := d.GetMinerId(userId)
		if minerId == 0 {
			d.sleepTime = sleep
		}
	}
	return nil
}

func ConfigInit() {
	// мониторим config.ini на наличие изменений
	go func() {
		for {
			log.Debug("ConfigInit monitor")
			if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
				utils.Sleep(1)
				continue
			}
			configIni_, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			configIni, err = configIni_.GetSection("default")
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			if len(configIni["db_type"]) > 0 {
				break
			}
			utils.Sleep(3)
		}
	}()
}

func init() {
	flag.Parse()

}

func CheckDaemonsRestart(chBreaker chan bool, chAnswer chan string, goRoutineName string) bool {
	log.Debug("CheckDaemonsRestart %v %v", goRoutineName, utils.Caller(2))
	select {
	case <-chBreaker:
		log.Debug("DaemonCh true %v", goRoutineName)
		chAnswer <- goRoutineName
		return true
	default:
	}
	return false
}

func DbConnect(chBreaker chan bool, chAnswer chan string, goRoutineName string) *utils.DCDB {
	for {
		if CheckDaemonsRestart(chBreaker, chAnswer, goRoutineName) {
			return nil
		}
		if utils.DB == nil || utils.DB.DB == nil {
			utils.Sleep(1)
		} else {
			return utils.DB
		}
	}
	return nil
}
