package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/mgo.v2/bson"
	"github.com/mgutz/ansi"
	"os/signal"
	"syscall"
	"log"
	"os/exec"
)

var host string
var fileOutput string
var waitForSignal bool
var db string

var blueFormat func(string) string = ansi.ColorFunc("blue+b+h")
var greenFormat func(string) string = ansi.ColorFunc("green+b+h")
var redFormat func(string) string = ansi.ColorFunc("red+b+h")
var resetFormat string = ansi.ColorCode("reset")

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "host to connect to, defaults to localhost")
	flag.StringVar(&fileOutput, "fileOutput", "setup", "Prefix to use for all files to use as target dump of the setup script")
	flag.BoolVar(&waitForSignal, "waitForSignal", true, "should program wait for Ctrl+Z before it fetches changes from DB?")
	flag.StringVar(&db, "db", "test", "Which DB to monitor")
	flag.Parse()
}

type CollectionItem struct {
	Id bson.ObjectId "_id"
}

type CollectionIds struct {
	Ids map[bson.ObjectId]bool
}

type Data map[string]CollectionIds



func main() {
	defer fmt.Println(resetFormat)

	context := Context{}
	context.connect()
	defer context.close()

	// TODO: test only
	run("/usr/local/bin/mongo", "localhost:27017/" + db, "/tmp/test_pre.js")

	beforeData := context.collectData()

	waitForStop()

	// TODO: test only
	run("/usr/local/bin/mongo", "localhost:27017/" + db, "/tmp/test.js")

	afterData := context.collectData()

	diffData := context.diffData(beforeData, afterData)

	if len(diffData) == 0 {
		fmt.Println(redFormat("No changes detected!"))
	} else {
		context.presentDiffData(diffData)
		context.makeJavaScript(fileOutput, diffData)
	}

	run("chmod", "+x", "./setup.sh")
	run("/bin/bash", "./setup.sh")

	diffData = context.diffData(afterData, context.collectData())

	if len(diffData) == 0 {
		fmt.Println(redFormat("No changes detected!"))
		return
	} else {
		context.presentDiffData(diffData)
	}
}

func waitForStop() {
	done := make(chan bool)

	go func () {
		fmt.Println(blueFormat("Send SIGTSTP (Ctrl+Z) when completed the introduction of things you wish to put in demo contents"))
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, syscall.SIGTSTP)
		select {
		case c:= <- signalChannel:
			fmt.Println(blueFormat("Signal received: ") + redFormat(c.String()))
			done<-true
		}
	}()

	if !waitForSignal {
		return
	}
	<-done
}


//TODO: just for testing

func run(name string, args ...string) {
	//TODO: don't keep when done
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		log.Fatalln("Failed when running", name, "with args", args, ":", err, "output:", string(out))
	}
	fmt.Println(greenFormat(fmt.Sprintf("Output of command %s with args %s %s\n", name, args, out)))
}

