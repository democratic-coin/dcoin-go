// +build !ios,!android

package main

import (
	"github.com/c-darwin/dcoin-go/packages/dcoin"
	"github.com/c-darwin/go-thrust/thrust"
	"github.com/c-darwin/go-thrust/lib/commands"
	"github.com/c-darwin/dcoin-go/packages/static"
	"fmt"
	"net/http"
	"github.com/c-darwin/go-thrust/lib/bindings/window"
	"os"
	"github.com/c-darwin/dcoin-go/packages/utils"
	"runtime"
)
func main_loader(w http.ResponseWriter, r *http.Request) {
	data, _ := static.Asset("static/img/main_loader.gif")
	fmt.Fprint(w, string(data))
}
func main_loader_html(w http.ResponseWriter, r *http.Request) {
	html := `<html><title>Dcoin</title><body style="margin:0;padding:0"><img src="static/img/main_loader.gif"/></body></html>`
	fmt.Fprint(w, html)
}
func main() {

	runtime.LockOSThread()

	var width uint = 800
	var height uint = 600
	var thrustWindow *window.Window
	thrust_shell := "thrust_shell"
	if runtime.GOOS == "windows" {
		thrust_shell = "thrust_shell.exe"
	} else if runtime.GOOS == "darwin" {
		thrust_shell = "ThrustShell"
		height = 578
	}
	if _, err := os.Stat(*utils.Dir+"/"+thrust_shell); err == nil && (winVer() >= 6|| winVer() == 0) {
		thrust.InitLogger()
		thrust.Start()
		thrustWindow = thrust.NewWindow(thrust.WindowOptions{
			RootUrl:  "http://localhost:8989/loader.html",
			HasFrame: true,
			Title : "Dcoin",
			Size: commands.SizeHW{Width:width, Height:height},
		})

		thrust.NewEventHandler("*", func(cr commands.CommandResponse) {
			fmt.Println(fmt.Sprintf("======Event(%s %d) - Signaled by Command (%s)", cr.TargetID, cr.Type))
			if cr.TargetID > 1 && cr.Type == "closed" {
				if utils.DB != nil && utils.DB.DB != nil {
					utils.DB.ExecSql(`INSERT INTO stop_daemons(stop_time) VALUES (?)`, utils.Time())
				} else {
					thrust.Exit()
					os.Exit(0)
				}
			}
		})
		thrustWindow.Show()
		thrustWindow.Focus()
		go func() {
			http.HandleFunc("/static/img/main_loader.gif", main_loader)
			http.HandleFunc("/loader.html", main_loader_html)
			http.ListenAndServe(":8989", nil)
		}()
	}
	tray()

	go dcoin.Start("", thrustWindow)

	enterLoop()

}