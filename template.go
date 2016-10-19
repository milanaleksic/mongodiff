package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"text/template"
	"runtime"
)

type templateData struct {
	DbName            string
	Filename          string
	Username          string
	Password          string
	CollectionChanges []collectionChange
}

type collectionChange struct {
	CollectionName   string
	ImportScriptName string
	AddedIds         []string
}

type templateConfiguration struct {
	filenamePattern string
	templateName    string
	mode            os.FileMode
}

var templateConfigurations = []templateConfiguration{
	{"{{.Filename}}_clean.js", "data/template_js", 0600},
	{"{{.Filename}}.bat", "data/template_replay_bat", 0600},
	{"{{.Filename}}.sh", "data/template_replay_bash", 0700},
	{"{{.Filename}}_clean.bat", "data/template_clean_bat", 0600},
	{"{{.Filename}}_clean.sh", "data/template_clean_bash", 0700},
}

func (templateData *templateData) WriteTemplates() {
	var toRemove []*os.File
	var toFlush []*bufio.Writer
	defer func() {
		for _, f := range toFlush {
			_ = f.Flush()
		}
		for _, f := range toRemove {
			_ = f.Close()
		}
	}()
	for _, configuration := range templateConfigurations {
		file := openFileOrFatal(templateData.getFilename(configuration))
		if runtime.GOOS != "windows" {
			if err := file.Chmod(configuration.mode); err != nil {
				log.Printf("Could not change file privileges for file %s, err:%v", file.Name(), err)
			}
		}
		toRemove = append(toRemove, file)
		fileWriter := bufio.NewWriter(file)
		toFlush = append(toFlush, fileWriter)
		template, err := template.New(configuration.templateName).Parse(string(MustAsset(configuration.templateName)))
		if err != nil {
			log.Fatalf("Template couldn't be parsed %s, err:%v", configuration.filenamePattern, err)
		}
		if err := template.Execute(fileWriter, templateData); err != nil {
			log.Fatalf("Template couldn't be expanded %s, err:%v", configuration.filenamePattern, err)
		}
	}
}

func (templateData *templateData) getFilename(configuration templateConfiguration) string {
	template, err := template.New(configuration.filenamePattern).Parse(configuration.filenamePattern)
	if err != nil {
		log.Fatalf("Template couldn't be parsed %s, err:%v", configuration.filenamePattern, err)
	}
	var filenameBuffer = &bytes.Buffer{}
	if err := template.Execute(filenameBuffer, templateData); err != nil {
		log.Fatalf("Template couldn't be expanded %s, err:%v", configuration.filenamePattern, err)
	}
	return string(filenameBuffer.Bytes())
}
