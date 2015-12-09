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
	havingMongoServerRunningInTheBackground(t)
	preFile := havingTestDataRemovalScript(t)
	defer os.Remove(preFile.Name())
	postFile := havingTestDataInjectionScript(t)
	defer os.Remove(postFile.Name())
	context := havingContextInstance()
	defer context.close()

	diffData := thenCalculationOfDeltaContains(t, context,
		func() { run("mongo", "localhost:27017/test", preFile.Name()) },
		func() { run("mongo", "localhost:27017/test", postFile.Name()) },
		[]interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")},
	)

	context.makeScriptFiles(diffData)
	defer os.Remove("./testing_clean.js")
	defer os.Remove("./testing_diffTest.json")
	defer os.Remove("./testing.sh")
	defer os.Remove("./testing.bat")
	defer os.Remove("./testing_clean.sh")
	defer os.Remove("./testing_clean.bat")

	thenDiffJsonHasExpectedChange(t, `"_id":{"$oid":"501ca04b668d67b3d6489f3a"}`)

	thenCalculationOfDeltaContains(t, context, func() {
		os.Setenv("MONGO_SERVER", "localhost")
		run("mongo", "localhost:27017/test", preFile.Name())
	}, func() {
		if runtime.GOOS == "windows" {
			run("testing_clean.bat")
			run("testing.bat")
		} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			run("chmod", "+x", "./testing_clean.sh")
			run("/bin/bash", "./testing_clean.sh")
			run("chmod", "+x", "./testing.sh")
			run("/bin/bash", "./testing.sh")
		} else {
			t.Error("Can't complete test since this platform is not supported: only linux and windows are supported")
			t.FailNow()
		}
	}, []interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")})
}

func TestDiffDetectionViaParameter(t *testing.T) {
	havingMongoServerRunningInTheBackground(t)
	preFile := havingTestDataRemovalScript(t)
	defer os.Remove(preFile.Name())
	postFile := havingTestDataInjectionScript(t)
	defer os.Remove(postFile.Name())
	context := havingContextInstance()
	defer context.close()

	diffData := thenCalculationOfDeltaContains(t, context,
		func() { run("mongo", "localhost:27017/test", preFile.Name()) },
		func() { run("mongo", "localhost:27017/test", postFile.Name()) },
		[]interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")},
	)

	context.makeScriptFiles(diffData)
	defer os.Remove("./testing_clean.js")
	defer os.Remove("./testing_diffTest.json")
	defer os.Remove("./testing.sh")
	defer os.Remove("./testing.bat")
	defer os.Remove("./testing_clean.sh")
	defer os.Remove("./testing_clean.bat")

	thenDiffJsonHasExpectedChange(t, `"_id":{"$oid":"501ca04b668d67b3d6489f3a"}`)

	thenCalculationOfDeltaContains(t, context, func() {
		run("mongo", "localhost:27017/test", preFile.Name())
	}, func() {
		if runtime.GOOS == "windows" {
			run("testing_clean.bat", "localhost")
			run("testing.bat", "localhost")
		} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			run("chmod", "+x", "./testing_clean.sh", "localhost")
			run("/bin/bash", "./testing_clean.sh", "localhost")
			run("chmod", "+x", "./testing.sh", "localhost")
			run("/bin/bash", "./testing.sh", "localhost")
		} else {
			t.Error("Can't complete test since this platform is not supported: only linux and windows are supported")
			t.FailNow()
		}
	}, []interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")})
}

func TestDiffDetectionBug5(t *testing.T) {
	havingMongoServerRunningInTheBackground(t)
	preFile := havingTestDataRemovalScriptForBug5(t)
	defer os.Remove(preFile.Name())
	postFile := havingTestDataInjectionScriptForBug5(t)
	defer os.Remove(postFile.Name())
	context := havingContextInstance()
	defer context.close()

	diffData := thenCalculationOfDeltaContains(t, context,
		func() { run("mongo", "localhost:27017/test", preFile.Name()) },
		func() { run("mongo", "localhost:27017/test", postFile.Name()) },
		[]interface{}{"foo"},
	)

	context.makeScriptFiles(diffData)
	defer os.Remove("./testing_clean.js")
	defer os.Remove("./testing_diffTest.json")
	defer os.Remove("./testing.sh")
	defer os.Remove("./testing.bat")
	defer os.Remove("./testing_clean.sh")
	defer os.Remove("./testing_clean.bat")

	thenDiffJsonHasExpectedChange(t, `{"_id":"foo"}`)

	thenCalculationOfDeltaContains(t, context, func() {
		os.Setenv("MONGO_SERVER", "localhost")
		run("mongo", "localhost:27017/test", preFile.Name())
	}, func() {
		if runtime.GOOS == "windows" {
			run("testing_clean.bat")
			run("testing.bat")
		} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			run("chmod", "+x", "./testing_clean.sh")
			run("/bin/bash", "./testing_clean.sh")
			run("chmod", "+x", "./testing.sh")
			run("/bin/bash", "./testing.sh")
		} else {
			t.Error("Can't complete test since this platform is not supported: only linux and windows are supported")
			t.FailNow()
		}
	}, []interface{}{"foo"})
}

func run(name string, args ...string) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		fmt.Errorf("Failed when running", name, "with args", args, ":", err, "output:", string(out))
	}
	fmt.Println(greenFormat(fmt.Sprintf("Output of command %s with args %s %s\n", name, args, out)))
}

func havingTestDataRemovalScript(t *testing.T) (preFile *os.File) {
	preFile, err := ioutil.TempFile("", "test_pre")
	if err != nil {
		t.Error("Could not open temp file", err)
	}
	preFile.WriteString(`db.diffTest.remove({ "_id" : ObjectId("501ca04b668d67b3d6489f3a") });`)
	return
}

func havingTestDataRemovalScriptForBug5(t *testing.T) (preFile *os.File) {
	preFile, err := ioutil.TempFile("", "test_pre")
	if err != nil {
		t.Error("Could not open temp file", err)
	}
	preFile.WriteString(`db.diffTest.remove({ "_id" : "foo" });`)
	return
}

func havingTestDataInjectionScript(t *testing.T) (postFile *os.File) {
	postFile, err := ioutil.TempFile("", "test_post")
	if err != nil {
		t.Error("Could not open temp file", err)
	}
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
	return
}

func havingTestDataInjectionScriptForBug5(t *testing.T) (postFile *os.File) {
	postFile, err := ioutil.TempFile("", "test_post")
	if err != nil {
		t.Error("Could not open temp file", err)
	}
	postFile.WriteString(`
		db.diffTest.insert({"_id": "foo"});
	`)
	return
}

func havingMongoServerRunningInTheBackground(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:27017")
	if err != nil {
		t.Error("Test can't be executed without running Mongo process")
		t.FailNow()
	}
	defer conn.Close()
}

func thenCalculationOfDeltaContains(t *testing.T, context *Context, preHook func(), changeHook func(), changeIds []interface{}) (diffData Data) {
	preHook()
	beforeData := context.collectData()
	changeHook()
	afterData := context.collectData()
	diffData = context.diffData(beforeData, afterData)

	if len(diffData) != len(changeIds) {
		t.Error("Expected this many elements in diff:", len(changeIds), "but was:", len(diffData))
		t.FailNow()
	}
	if len(diffData["diffTest"].Ids) != len(changeIds) {
		t.Error("Expected this many elements in diff:", len(changeIds), "but was:", len(diffData["diffTest"].Ids))
		t.FailNow()
	}
	for _, id := range changeIds {
		if !diffData["diffTest"].Ids[id] {
			t.Error("Could not find expected ID in the diff!", len(diffData["diffTest"].Ids))
			t.FailNow()
		}
	}
	return
}

func thenDiffJsonHasExpectedChange(t *testing.T, contents string) {
	data, err := ioutil.ReadFile("./testing_diffTest.json")
	if err != nil {
		t.Error("Could not verify generated JSON file!", err)
		t.FailNow()
	}
	if !strings.Contains(string(data), contents) {
		t.Error("Could not verify generated JS file. Contents:", string(data), "expected:", contents)
		t.FailNow()
	}
}

func havingContextInstance() (context *Context) {
	context = &Context{
		dbName: "test",
		host:   "localhost",
		prefix: "testing",
	}
	context.connect()
	return
}