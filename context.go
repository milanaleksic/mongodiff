package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"strings"
	"text/template"
)

type CollectionItem struct {
	Id bson.ObjectId "_id"
}

type CollectionIds struct {
	Ids map[bson.ObjectId]bool
}

type Data map[string]CollectionIds

type Context struct {
	session  *mgo.Session
	db       *mgo.Database
	host     string
	dbName   string
	excludes string
}

func (context *Context) connect() {
	var err error
	context.session, err = mgo.Dial(context.host)
	if err != nil {
		panic(err)
	}
	context.session.SetMode(mgo.Monotonic, true)
	context.db = context.session.DB(context.dbName)
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
outer:
	for _, collection := range collections {
		for _, exclude := range strings.Split(context.excludes, ",") {
			if collection == exclude {
				continue outer
			}
		}
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
		log.Fatalf("Could not open file for saving: %s; error: %s", filename, err)
	}
	return
}

type TemplateData struct {
	DbName            string
	Filename          string
	CollectionChanges []CollectionChange
}

type CollectionChange struct {
	CollectionName   string
	ImportScriptName string
	AddedIds         []string
}

func (context *Context) makeScriptFiles(filename string, diffData Data) {
	templateData := TemplateData{
		DbName: context.dbName, Filename: filename,
	}

	for collectionName, ids := range diffData {
		importScriptFilename := fmt.Sprintf("%s_%s.json", filename, collectionName)
		importScript := openFileOrFatal(importScriptFilename)
		defer importScript.Close()
		writerImportScript := bufio.NewWriter(importScript)

		newIds := make([]string, 0)
		for id, _ := range ids.Ids {
			newIds = append(newIds, id.Hex())
			context.dumpJsonToFile(collectionName, id, writerImportScript)
		}
		templateData.CollectionChanges = append(templateData.CollectionChanges, CollectionChange{
			collectionName, importScriptFilename, newIds,
		})
	}

	writeTemplate(&templateData, filename+"_clean.js", "data/template_js", 0600)
	writeTemplate(&templateData, filename+".bat", "data/template_bat", 0600)
	writeTemplate(&templateData, filename+".sh", "data/template_bash", 0700)
}

func (context *Context) dumpJsonToFile(collectionName string, id bson.ObjectId, writerImportScript *bufio.Writer) {
	raw := bson.D{}
	err := context.db.C(collectionName).Find(bson.M{"_id": id}).One(&raw)
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

func writeTemplate(templateData *TemplateData, filename string, templateName string, mode os.FileMode) {
	file := openFileOrFatal(filename)
	file.Chmod(mode)
	defer file.Close()
	fileWriter := bufio.NewWriter(file)
	defer fileWriter.Flush()
	template, err := template.New(templateName).Parse(string(MustAsset(templateName)))
	if err != nil {
		log.Fatalf("Template couldn't be parsed", err)
	}
	template.Execute(fileWriter, templateData)
}
