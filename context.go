package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"os"
	"strings"
)

type CollectionItem struct {
	Id interface{} "_id"
}

type CollectionIds struct {
	Ids map[interface{}]bool
}

type Data map[string]CollectionIds

type Context struct {
	session  *mgo.Session
	db       *mgo.Database
	host     string
	dbName   string
	excludes string
	prefix   string
}

func (context *Context) checkMongoUp() (err error) {
	target := context.host
	fmt.Println(target)
	if !strings.Contains(target, ":") {
		target = target + ":27017"
	}
	fmt.Printf("Checking if Mongo is up... %s\n", target)
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return err
	}
	defer conn.Close()
	return
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
		ids := make(map[interface{}]bool, 0)
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
	// fmt.Printf("Before: %v\n After: %v", before, after)
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
						Ids: make(map[interface{}]bool, 0),
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
			fmt.Println("\t\t", greenFormat(fmt.Sprintf("%v", id)))
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

func (context *Context) makeScriptFiles(diffData Data) {
	templateData := TemplateData{
		DbName: context.dbName, Filename: context.prefix,
	}

	for collectionName, ids := range diffData {
		importScriptFilename := fmt.Sprintf("%s_%s.json", context.prefix, collectionName)
		importScript := openFileOrFatal(importScriptFilename)
		defer importScript.Close()
		writerImportScript := bufio.NewWriter(importScript)

		newIds := make([]string, 0)
		for id, _ := range ids.Ids {
			switch t := id.(type) {
			case bson.ObjectId:
				newIds = append(newIds, fmt.Sprintf(`ObjectId("%v")`, id.(bson.ObjectId).Hex()))
			case string:
				newIds = append(newIds, fmt.Sprintf(`"%v"`, id.(string)))
			default:
				log.Fatalf("Can not handle this type: [%T] yet, please report issue on github.com/milanaleksic/mongodiff", t)
			}
			context.dumpJsonToFile(collectionName, id, writerImportScript)
		}
		templateData.CollectionChanges = append(templateData.CollectionChanges, CollectionChange{
			collectionName, importScriptFilename, newIds,
		})
	}
	templateData.WriteTemplates()
}

func (context *Context) dumpJsonToFile(collectionName string, id interface{}, writerImportScript *bufio.Writer) {
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
