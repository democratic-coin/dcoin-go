// desktoplite
package main

import (
	"fmt"
	"net/http"
	"github.com/democratic-coin/dcoin-go/packages/system"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"os"
	"runtime"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

type Pool struct {
	Pool string `json:"pool"`
}

func main() {
	var ( thrustWindow *window.Window
		pool Pool
	)

	dir,_ := filepath.Abs(filepath.Dir(os.Args[0]))
	userfile := filepath.Join(dir, `iduser.txt`)
	txtUser,_ := ioutil.ReadFile(userfile)
	idUser := utils.StrToInt64(string(txtUser))
	resp, err := http.Get(`http://getpool.dcoin.club/?user_id=` + utils.Int64ToStr(idUser))
	if err!=nil {
		os.Exit(1)
	}
	jsonPool, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err!=nil {
		os.Exit(1)
	}
	json.Unmarshal(jsonPool, &pool)
	if pool.Pool == `0` || len(pool.Pool) == 0 {
		pool.Pool = `http://pool.dcoin.club`
	}
	
	fmt.Println( pool.Pool )
	
	runtime.LockOSThread()
	//	if utils.Desktop() && (winVer() >= 6 || winVer() == 0) {
	thrust.Start()

	thrust.NewEventHandler("*", func(cr commands.CommandResponse) {
/*		cr_marshaled, err := json.Marshal(cr)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(fmt.Sprintf("Event(%s) - Signaled by Command (%s)", cr.Type, cr_marshaled))
		}*/
		if cr.Type == "closed" {
			system.FinishThrust(0)
			os.Exit(0)
		}
	})
	
	thrustWindow = thrust.NewWindow(thrust.WindowOptions{
		RootUrl: pool.Pool,
		Size:    commands.SizeHW{Width: 1024, Height: 600},
	})
/*	thrustWindow.HandleEvent("*", func(cr commands.EventResult) {
		fmt.Println("HandleEvent", cr)
	})*/
	thrustWindow.HandleRemote(func(er commands.EventResult, this *window.Window) {
		fmt.Println("RemoteMessage Recieved:", er.Message.Payload)
		if er.Message.Payload[:7]==`USERID=` {
			err := ioutil.WriteFile( userfile, []byte(er.Message.Payload[7:]), 0644 )
			if err != nil {
				fmt.Println( `Error`, err )
			}
		} else {
			utils.ShellExecute(er.Message.Payload)
		}
	})
	
	thrustWindow.Show()
	thrustWindow.Focus()
//	thrustWindow.OpenDevtools()
	for {
		utils.Sleep(3600)
	}
	system.Finish(0)
}
