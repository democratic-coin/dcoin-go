package stopdaemons

import (
	"fmt"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"github.com/c-darwin/dcoin-go/vendor/src/github.com/c-darwin/go-thrust/thrust"
	"github.com/c-darwin/dcoin-go/vendor/src/github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger("stop_daemons")

func WaitStopTime() {
	var first bool
	for {
		if utils.DB == nil || utils.DB.DB == nil {
			utils.Sleep(3)
			continue
		}
		if !first {
			err := utils.DB.ExecSql(`DELETE FROM stop_daemons`)
			if err != nil {
				log.Error(utils.ErrInfo(err).Error())
			}
			first = true
		}
		dExists, err := utils.DB.Single(`SELECT stop_time FROM stop_daemons`).Int64()
		if err != nil {
			log.Error(utils.ErrInfo(err).Error())
		}
		log.Debug("dExtit: %d", dExists)
		if dExists > 0 {
			fmt.Println("Stop_daemons from DB!")
			for _, ch := range utils.DaemonsChans {
				fmt.Println("ch.ChBreaker<-true")
				ch.ChBreaker<-true
			}
			for _, ch := range utils.DaemonsChans {
				fmt.Println(<-ch.ChAnswer)
			}
			fmt.Println("Daemons killed")
			err := utils.DB.Close()
			if err != nil {
				log.Error(utils.ErrInfo(err).Error())
			}
			fmt.Println("DB Closed")
			err = os.Remove(*utils.Dir + "/dcoin.pid")
			if err != nil {
				log.Error(utils.ErrInfo(err).Error())
				panic(err)
			}
			fmt.Println("removed " + *utils.Dir + "/dcoin.pid")
			thrust.Exit()
			os.Exit(1)
		}
		utils.Sleep(1)
	}
}
