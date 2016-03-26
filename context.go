package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type collectionItem struct {
	ID interface{} `bson:"_id"`
}

type collectionIds struct {
	Ids map[interface{}]bool
}

type data map[string]collectionIds

type context struct {
	session  *mgo.Session
	db       *mgo.Database
	host     string
	dbName   string
	excludes string
	prefix   string
}

func (context *context) checkMongoUp() (err error) {
	target := context.host
	if !strings.Contains(target, ":") {
		target = target + ":27017"
	}
	fmt.Printf("Checking if Mongo is up... %s\n", target)
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	return
}

func (context *context) connect() {
	var err error
	context.session, err = mgo.Dial(context.host)
	if err != nil {
		panic(err)
	}
	context.session.SetMode(mgo.Monotonic, true)
	context.db = context.session.DB(context.dbName)
}

func (context *context) close() {
	context.session.Close()
}

func (context *context) collectData() (collectedData data) {
	var maxLength = 0
	defer func() {
		fmt.Printf("\rScanning completed!%*s\n", maxLength+1, "")
	}()

	collections, err := context.db.CollectionNames()
	if err != nil {
		log.Fatal("Could not fetch collection names (?)", err)
	}
	collectedData = make(data)

outer:
	for _, collection := range collections {
		if ln := len(collection); ln > maxLength {
			maxLength = ln
		}
		fmt.Printf("\r%sScanning collection %s", resetFormat, redFormat(collection))
		for _, exclude := range strings.Split(context.excludes, ",") {
			if collection == exclude {
				continue outer
			}
		}
		ids := make(map[interface{}]bool, 0)
		iter := context.db.C(collection).Find(nil).Iter()

		collItem := collectionItem{}
		for iter.Next(&collItem) {
			// fmt.Printf("Result: %v\n", collItem.Id)
			ids[collItem.ID] = true
		}
		if err := iter.Close(); err != nil {
			log.Fatal("Could not close iterator", err)
		}

		collectedData[collection] = collectionIds{
			Ids: ids,
		}
	}
	return
}

func (context *context) diffData(before data, after data) data {
	changes := data{}
	// fmt.Printf("Before: %v\n After: %v", before, after)
	for collectionName, knownIds := range before {
		newItems, ok := after[collectionName]
		if !ok {
			changes[collectionName] = knownIds
			continue
		}
		for maybeANewID := range newItems.Ids {
			if _, ok := knownIds.Ids[maybeANewID]; !ok {
				if _, ok := changes[collectionName]; !ok {
					changes[collectionName] = collectionIds{
						Ids: make(map[interface{}]bool, 0),
					}
				}
				changes[collectionName].Ids[maybeANewID] = true
			}
		}
	}
	for collectionName := range after {
		known, ok := before[collectionName]
		if !ok {
			changes[collectionName] = known
		}
	}
	return changes
}

func (context *context) presentDiffData(diffData data) {
	fmt.Println(redFormat("All changed data: "))
	for collectionName, ids := range diffData {
		fmt.Println("\t", blueFormat(collectionName))
		for id := range ids.Ids {
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

func (context *context) makeScriptFiles(diffData data) {
	templateData := templateData{
		DbName: context.dbName, Filename: context.prefix,
	}

	var toRemove []*os.File
	defer func() {
		for _, f := range toRemove {
			_ = f.Close()
		}
	}()

	for collectionName, ids := range diffData {
		importScriptFilename := fmt.Sprintf("%s_%s.json", context.prefix, collectionName)
		importScript := openFileOrFatal(importScriptFilename)
		toRemove = append(toRemove, importScript)
		writerImportScript := bufio.NewWriter(importScript)

		var newIds []string
		for id := range ids.Ids {
			switch t := id.(type) {
			case bson.ObjectId:
				newIds = append(newIds, fmt.Sprintf(`ObjectId("%v")`, id.(bson.ObjectId).Hex()))
			case string:
				newIds = append(newIds, fmt.Sprintf(`"%v"`, id.(string)))
			default:
				log.Fatalf("Can not handle this type: [%T] yet, please report issue on github.com/milanaleksic/mongodiff", t)
			}
			context.dumpJSONToFile(collectionName, id, writerImportScript)
		}
		templateData.CollectionChanges = append(templateData.CollectionChanges, collectionChange{
			collectionName, importScriptFilename, newIds,
		})
	}
	templateData.WriteTemplates()
}

func (context *context) dumpJSONToFile(collectionName string, id interface{}, writerImportScript *bufio.Writer) {
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
	err = writerImportScript.Flush()
	if err != nil {
		log.Fatalln("Could not flush the file contents!", err)
	}
}
