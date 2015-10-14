package main

import (
	mgo "gopkg.in/mgo.v2"
	"bufio"
	"fmt"
	"os"
	"log"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"github.com/mongodb/mongo-tools/common/bsonutil"
)

type Context struct {
	session *mgo.Session
	db      *mgo.Database
}

func (context *Context) connect() {
	var err error
	context.session, err = mgo.Dial(host)
	if err != nil {
		panic(err)
	}
	context.session.SetMode(mgo.Monotonic, true)
	context.db = context.session.DB(db)
}

func (context *Context) close() {
	context.session.Close()
}

func (context *Context) collectData() (data Data) {
	collections, err := context.db.CollectionNames()
	if err != nil {
		log.Fatal("Could not fetch collection names (?)", err)
	}
	data = make(Data)
	for _, collection := range collections {
		ids := make(map[bson.ObjectId]bool, 0)
		iter := context.db.C(collection).Find(nil).Iter()

		collectionItem := CollectionItem{}
		for iter.Next(&collectionItem) {
			// fmt.Printf("Result: %v\n", collectionItem.Id)
			ids[collectionItem.Id] = true
		}
		if err := iter.Close(); err != nil {
			log.Fatal("Could not close iterator", err)
		}

		data[collection] = CollectionIds{
			Ids: ids,
		}
	}
	return
}

func (context *Context) diffData(before Data, after Data) Data {
	changes := Data{}
	for collectionName, knownIds := range before {
		newItems, ok := after[collectionName]
		if !ok {
			changes[collectionName] = knownIds
			continue
		}
		for maybeANewId, _ := range newItems.Ids {
			if _, ok := knownIds.Ids[maybeANewId]; !ok {
				if _, ok := changes[collectionName]; !ok {
					changes[collectionName] = CollectionIds{
						Ids: make(map[bson.ObjectId]bool, 0),
					}
				}
				changes[collectionName].Ids[maybeANewId] = true
			}
		}
	}
	for collectionName, _ := range after {
		known, ok := before[collectionName]
		if !ok {
			changes[collectionName] = known
		}
	}
	return changes
}

func (context *Context) presentDiffData(diffData Data) {
	fmt.Println(redFormat("All changed data: "))
	for collectionName, ids := range diffData {
		fmt.Println("\t", blueFormat(collectionName))
		for id, _ := range ids.Ids {
			fmt.Println("\t\t", greenFormat(id.Hex()))
		}
	}
}

func openFileOrFatal(filename string) (file *os.File) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not open file for saving", filename)
	}
	return
}

func (context *Context) makeJavaScript(filename string, diffData Data) {
	script := openFileOrFatal(filename + ".sh")
	defer script.Close()
	writerScript := bufio.NewWriter(script)

	javaScript := openFileOrFatal(filename + ".js")
	defer javaScript.Close()
	writerJavaScript := bufio.NewWriter(javaScript)

	fmt.Fprintln(writerScript, "#!/bin/bash")
	fmt.Fprintln(writerScript, fmt.Sprintf("mongo $MONGO_SERVER/%s %s.js", db, filename))

	for collectionName, ids := range diffData {
		importScriptFilename := fmt.Sprintf("%s_%s.json", filename, collectionName)
		fmt.Fprintln(writerScript, fmt.Sprintf("mongoimport --host $MONGO_SERVER --db %s --collection %s < %s", db, collectionName, importScriptFilename))
		importScript := openFileOrFatal(importScriptFilename)
		defer importScript.Close()
		writerImportScript := bufio.NewWriter(importScript)

		for id, _ := range ids.Ids {
			fmt.Fprintf(writerJavaScript, "db.%s.remove({\"_id\":ObjectId(\"%s\")});\n", collectionName, id.Hex())

			context.dumpJsonToFile(collectionName, id, writerImportScript)
		}
	}
	writerScript.Flush()
	writerJavaScript.Flush()
}

func (context *Context) dumpJsonToFile(collectionName string, id bson.ObjectId, writerImportScript *bufio.Writer) {
	raw := bson.D{}
	err := context.db.C(collectionName).Find(bson.M{"_id":id}).One(&raw)
	if err != nil {
		log.Fatalln("Could not open Marshal raw data", err)
	}
	output, err := bsonutil.ConvertBSONValueToJSON(raw)
	if err != nil {
		log.Fatalln("Could not convert to JSON", err)
	}
	var out []byte
	out, err = json.Marshal(output)
	if err != nil {
		log.Fatalln("Could not Marshal JSON into bytes", err)
	}
	fmt.Fprintf(writerImportScript, "%s\n", out)
	writerImportScript.Flush()
}