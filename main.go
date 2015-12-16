package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mgutz/ansi"
	"os/signal"
)

var blueFormat func(string) string = ansi.ColorFunc("blue+b+h")
var greenFormat func(string) string = ansi.ColorFunc("green+b+h")
var redFormat func(string) string = ansi.ColorFunc("red+b+h")
var resetFormat string = ansi.ColorCode("reset")

func main() {
	defer fmt.Println(resetFormat)

	var host *string = flag.String("host", "127.0.0.1", "host to connect to, defaults to localhost")
	var fileOutput *string = flag.String("fileOutput", "setup", "Prefix to use for all files to use as target dump of the setup script")
	var waitForSignal *bool = flag.Bool("waitForSignal", true, "should program wait for Ctrl+C before it fetches changes from DB?")
	var dbName *string = flag.String("db", "test", "Which DB to monitor")
	var excludes *string = flag.String("excludes", "", "Which collections to ignore")
	flag.Parse()

	context := Context{
		host:     *host,
		dbName:   *dbName,
		excludes: *excludes,
		prefix:   *fileOutput,
	}
	if err := context.checkMongoUp(); err != nil {
		fmt.Println("Mongo is not up!")
		os.Exit(1)
	}
	context.connect()
	defer context.close()

	beforeData := context.collectData()

	waitForStop(*waitForSignal)

	afterData := context.collectData()

	diffData := context.diffData(beforeData, afterData)

	if len(diffData) == 0 {
		fmt.Println(redFormat("No changes detected!"))
	} else {
		context.presentDiffData(diffData)
		context.makeScriptFiles(diffData)
	}
}

func waitForStop(waitForSignal bool) {
	done := make(chan bool)

	go func() {
		fmt.Println(blueFormat("Send SIGINT (") + redFormat("Ctrl+C") + blueFormat(") when completed the introduction of things you wish to put in demo contents"))
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, os.Interrupt)
		select {
		case c := <-signalChannel:
			fmt.Println(blueFormat("Signal received: ") + redFormat(c.String()))
			done <- true
		}
	}()

	if !waitForSignal {
		return
	}
	<-done
}
