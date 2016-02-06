package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestDiffDetection(t *testing.T) {
	preFile := havingTestDataRemovalScript(t)
	postFile := havingTestDataInjectionScript(t)
	defer removeTestFilesIncluding(preFile, postFile)

	context := havingContextInstance(t)
	defer context.close()

	diffData := thenCalculationOfDeltaContains(t, context,
		func() { run("mongo", "localhost:27017/test", preFile.Name()) },
		func() { run("mongo", "localhost:27017/test", postFile.Name()) },
		[]interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")},
	)

	context.makeScriptFiles(diffData)

	thenDIFFJsonHasExpectedChange(t, `"_id":{"$oid":"501ca04b668d67b3d6489f3a"}`)

	thenCalculationOfDeltaContains(t, context, func() {
		if err := os.Setenv("MONGO_SERVER", "localhost"); err != nil {
			t.Fatal("Could not set environment parameter!")
		}
		run("mongo", "localhost:27017/test", preFile.Name())
	}, func() { executeClean(t) }, []interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")})
}

func TestDiffDetectionViaParameter(t *testing.T) {
	preFile := havingTestDataRemovalScript(t)
	postFile := havingTestDataInjectionScript(t)
	defer removeTestFilesIncluding(preFile, postFile)

	context := havingContextInstance(t)
	defer context.close()

	diffData := thenCalculationOfDeltaContains(t, context,
		func() { run("mongo", "localhost:27017/test", preFile.Name()) },
		func() { run("mongo", "localhost:27017/test", postFile.Name()) },
		[]interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")},
	)

	context.makeScriptFiles(diffData)

	thenDIFFJsonHasExpectedChange(t, `"_id":{"$oid":"501ca04b668d67b3d6489f3a"}`)

	thenCalculationOfDeltaContains(t, context, func() {
		run("mongo", "localhost:27017/test", preFile.Name())
	}, func() {
		if runtime.GOOS == "windows" {
			run("testing_clean.bat", "localhost")
			run("testing.bat", "localhost")
		} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			run("chmod", "+x", "./testing_clean.sh")
			run("/bin/bash", "./testing_clean.sh", "localhost")
			run("chmod", "+x", "./testing.sh")
			run("/bin/bash", "./testing.sh", "localhost")
		} else {
			t.Error("Can't complete test since this platform is not supported: only linux and windows are supported")
			t.FailNow()
		}
	}, []interface{}{bson.ObjectIdHex("501ca04b668d67b3d6489f3a")})
}

func TestDiffDetectionBug5(t *testing.T) {
	preFile := havingTestDataRemovalScriptForBug5(t)
	postFile := havingTestDataInjectionScriptForBug5(t)
	defer removeTestFilesIncluding(preFile, postFile)

	context := havingContextInstance(t)
	defer context.close()

	diffData := thenCalculationOfDeltaContains(t, context,
		func() { run("mongo", "localhost:27017/test", preFile.Name()) },
		func() { run("mongo", "localhost:27017/test", postFile.Name()) },
		[]interface{}{"foo"},
	)

	context.makeScriptFiles(diffData)

	thenDIFFJsonHasExpectedChange(t, `{"_id":"foo"}`)

	thenCalculationOfDeltaContains(t, context, func() {
		if err := os.Setenv("MONGO_SERVER", "localhost"); err != nil {
			t.Fatal("Could not set environment parameter!")
		}
		run("mongo", "localhost:27017/test", preFile.Name())
	}, func() { executeClean(t) }, []interface{}{"foo"})
}

func executeClean(t *testing.T) {
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
}

func run(name string, args ...string) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		if err = fmt.Errorf("Failed when running %s with args %v: %v, output: %s", name, args, err, string(out)); err != nil {
			panic(fmt.Sprintf("Invalid string format: %v", err))
		}
	}
	fmt.Println(greenFormat(fmt.Sprintf("Output of command %s with args %s %s\n", name, args, out)))
}

func havingTestDataRemovalScript(t *testing.T) (preFile *os.File) {
	return testFile(t, "test_post", `db.diffTest.remove({ "_id" : ObjectId("501ca04b668d67b3d6489f3a") });`)
}

func havingTestDataRemovalScriptForBug5(t *testing.T) (preFile *os.File) {
	return testFile(t, "test_post", `db.diffTest.remove({ "_id" : "foo" });`)
}

func havingTestDataInjectionScript(t *testing.T) (postFile *os.File) {
	return testFile(t, "test_post", `
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
}

func havingTestDataInjectionScriptForBug5(t *testing.T) (postFile *os.File) {
	return testFile(t, "test_post", `db.diffTest.insert({"_id": "foo"});`)
}

func testFile(t *testing.T, prefix string, content string) (postFile *os.File) {
	postFile, err := ioutil.TempFile("", prefix)
	if err != nil {
		t.Error("Could not open temp file", err)
	}
	_, err = postFile.WriteString(content)
	if err != nil {
		t.Errorf("Could not write fully contents %s to temp file %s: %v", content, prefix, err)
	}
	return
}

func thenCalculationOfDeltaContains(t *testing.T, context *context, preHook func(), changeHook func(), changeIds []interface{}) (diffData data) {
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

func thenDIFFJsonHasExpectedChange(t *testing.T, contents string) {
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

func havingContextInstance(t *testing.T) (mongodiffContext *context) {
	mongodiffContext = &context{
		dbName: "test",
		host:   "localhost",
		prefix: "testing",
	}
	if err := mongodiffContext.checkMongoUp(); err != nil {
		t.Error("Test can't be executed without running Mongo process")
		t.FailNow()
	}
	mongodiffContext.connect()
	return
}

func removeTestFilesIncluding(extra ...*os.File) {
	_ = os.Remove("./testing_clean.js")
	_ = os.Remove("./testing_diffTest.json")
	_ = os.Remove("./testing.sh")
	_ = os.Remove("./testing.bat")
	_ = os.Remove("./testing_clean.sh")
	_ = os.Remove("./testing_clean.bat")
	for _, ex := range extra {
		_ = os.Remove(ex.Name())
	}
}
