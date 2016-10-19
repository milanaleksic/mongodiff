package main

import (
	"flag"
	"fmt"
	"os"

	"os/signal"

	"github.com/mgutz/ansi"
)

var blueFormat = ansi.ColorFunc("blue+b+h")
var greenFormat = ansi.ColorFunc("green+b+h")
var redFormat = ansi.ColorFunc("red+b+h")
var resetFormat = ansi.ColorCode("reset")

// Version holds the main version string which should be updated externally when building release
var Version = "undefined"

func main() {
	defer fmt.Println(resetFormat)

	var host = flag.String("host", "127.0.0.1", "host to connect to, defaults to localhost")
	var fileOutput = flag.String("fileOutput", "setup", "Prefix to use for all files to use as target dump of the setup script")
	var waitForSignal = flag.Bool("waitForSignal", true, "should program wait for Ctrl+C before it fetches changes from DB?")
	var dbName = flag.String("db", "test", "Which DB to monitor")
	var excludes = flag.String("excludes", "", "Which collections to ignore")
	var username = flag.String("username", "", "(Optional) which username to use to authenticate")
	var password = flag.String("password", "", "(Optional) which password to use to authenticate")
	var copyCredentials = flag.Bool("copyCredentials", false, "Should credentials be copied to the script")
	var version = flag.Bool("version", false, "Get application version")
	flag.Parse()

	if *version {
		fmt.Printf("mongodiff version: %v\n", Version)
		return
	}

	ctx := context{
		host:     *host,
		dbName:   *dbName,
		excludes: *excludes,
		prefix:   *fileOutput,
		username: *username,
		password: *password,
		copyCredentials: *copyCredentials,
	}
	if err := ctx.checkMongoUp(); err != nil {
		fmt.Println("Mongo is not up!")
		os.Exit(1)
	}
	ctx.connect()
	defer ctx.close()

	beforeData := ctx.collectData()

	waitForStop(*waitForSignal)

	afterData := ctx.collectData()

	diffData := ctx.diffData(beforeData, afterData)

	if len(diffData) == 0 {
		fmt.Println(redFormat("No changes detected!"))
	} else {
		ctx.presentDiffData(diffData)
		ctx.makeScriptFiles(diffData)
	}
}

func waitForStop(waitForSignal bool) {
	done := make(chan bool)

	go func() {
		fmt.Println(blueFormat("Send SIGINT (") + redFormat("Ctrl+C") + blueFormat(") when completed the introduction of things you wish to put in demo contents"))
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, os.Interrupt)
		c := <-signalChannel
		fmt.Println(blueFormat("Signal received: ") + redFormat(c.String()))
		done <- true
	}()

	if !waitForSignal {
		return
	}
	<-done
}
