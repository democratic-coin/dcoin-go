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
	var thrustWindow *window.Window
	thrust_shell := "thrust_shell"
	if runtime.GOOS == "windows" {
		thrust_shell = "thrust_shell.exe"
	}

	if _, err := os.Stat(*utils.Dir+"/"+thrust_shell); err == nil && (winVer() >= 6|| winVer( )== 0) {
		thrust.InitLogger()
		thrust.Start()
		thrustWindow = thrust.NewWindow(thrust.WindowOptions{
			RootUrl:  "http://localhost:8989/loader.html",
			HasFrame: true,
			Title : "Dcoin",
			Size: commands.SizeHW{Width:800, Height:600},
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

	dcoin.Start("", thrustWindow)
}