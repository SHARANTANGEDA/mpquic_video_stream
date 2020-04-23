package main

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os"
)

func main() {
	gtk.Init(nil)
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Video Stream Using QUIC Protocol")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	var cw ClientWorker
	fmt.Println("In Client")
	cw.clientWorker(win, os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk.Main()
}
