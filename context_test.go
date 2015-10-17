package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestDiffDetection(t *testing.T) {
	preFile, err := ioutil.TempFile("", "test_pre")
	if err != nil {
		t.Error("Could not open temp file", err)
	}
	defer os.Remove(preFile.Name())
	preFile.WriteString(`db.diffTest.remove({ "_id" : ObjectId("501ca04b668d67b3d6489f3a") });`)

	postFile, err := ioutil.TempFile("", "test_post")
	if err != nil {
		t.Error("Could not open temp file", err)
	}
	defer os.Remove(postFile.Name())
	postFile.WriteString(`
		db.diffTest.insert({
			"_id" : ObjectId("501ca04b668d67b3d6489f3a"),
			"a" : "aaf52b19-6c25-11e5-86ae-af5a85ef200c",
			"v" : {
				"c" : "!@$"
			},
			"array" : [
				{
					"asa" : [
						{
							"cas" : "235122",
							"array2" : [
								{
									"x" : 50.1,
									"a" : ISODate("2015-08-18T07:00:03.450+0000")
								}
							]
						}
					]
				}
			]
		});
	`)

	conn, err := net.Dial("tcp", "localhost:27017")
	if err != nil {
		t.Error("Test can't be executed without running Mongo process")
		t.FailNow()
	}
	conn.Close()

	context := Context{
		dbName: "test",
		host:   "localhost",
	}
	context.connect()
	defer context.close()

	run("mongo", "localhost:27017/test", preFile.Name())
	beforeData := context.collectData()
	run("mongo", "localhost:27017/test", postFile.Name())
	afterData := context.collectData()
	diffData := context.diffData(beforeData, afterData)

	if len(diffData) != 1 {
		t.Error("Expected one element in diff!", len(diffData))
		t.FailNow()
	}
	if len(diffData["diffTest"].Ids) != 1 {
		t.Error("Expected one Id in diff!", len(diffData["diffTest"].Ids))
		t.FailNow()
	}
	if !diffData["diffTest"].Ids[bson.ObjectIdHex("501ca04b668d67b3d6489f3a")] {
		t.Error("Could not find expected ID in the diff!", len(diffData["diffTest"].Ids))
		t.FailNow()
	}
	context.makeScriptFiles("testing", diffData)
	defer os.Remove("./testing.js")
	defer os.Remove("./testing_diffTest.json")
	defer os.Remove("./testing.sh")
	defer os.Remove("./testing.bat")

	data, err := ioutil.ReadFile("./testing_diffTest.json")
	if err != nil {
		t.Error("Could not verify generated JSON file!", err)
		t.FailNow()
	}
	if !strings.Contains(string(data), "$oid") {
		t.Error("Could not verify generated JS file. Contents:", string(data))
		t.FailNow()
	}

	os.Setenv("MONGO_SERVER", "localhost")
	run("mongo", "localhost:27017/test", preFile.Name())
	beforeAutoScript := context.collectData()
	if runtime.GOOS == "windows" {
		run("testing.bat")
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		run("chmod", "+x", "./testing.sh")
		run("/bin/bash", "./testing.sh")
	} else {
		t.Error("Can't complete test since this platform is not supported: only linux and windows are supported")
		t.FailNow()
	}

	afterAutoScript := context.collectData()

	diffData = context.diffData(beforeAutoScript, afterAutoScript)

	if len(diffData) != 1 {
		t.Error("Expected one element in diff!", len(diffData))
		t.FailNow()
	}
}

func run(name string, args ...string) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		fmt.Errorf("Failed when running", name, "with args", args, ":", err, "output:", string(out))
	}
	fmt.Println(greenFormat(fmt.Sprintf("Output of command %s with args %s %s\n", name, args, out)))
}
