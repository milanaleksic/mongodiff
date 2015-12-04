package main
import (
	"os"
	"bufio"
	"text/template"
	"log"
	"bytes"
)

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

type templateConfiguration struct {
	filenamePattern string
	templateName    string
	mode            os.FileMode
}

var templateConfigurations = []templateConfiguration {
	{"{{.Filename}}_clean.js", "data/template_js", 0600, },
	{"{{.Filename}}.bat", "data/template_replay_bat", 0600, },
	{"{{.Filename}}.sh", "data/template_replay_bash", 0700, },
	{"{{.Filename}}_clean.bat", "data/template_clean_bat", 0600, },
	{"{{.Filename}}_clean.sh", "data/template_clean_bash", 0700, },
}

func (templateData *TemplateData) WriteTemplates() {
	for _, configuration := range templateConfigurations {
		file := openFileOrFatal(templateData.getFilename(configuration))
		file.Chmod(configuration.mode)
		defer file.Close()
		fileWriter := bufio.NewWriter(file)
		defer fileWriter.Flush()
		template, err := template.New(configuration.templateName).Parse(string(MustAsset(configuration.templateName)))
		if err != nil {
			log.Fatalf("Template couldn't be parsed", err)
		}
		template.Execute(fileWriter, templateData)
	}
}

func (templateData *TemplateData) getFilename(configuration templateConfiguration) (string) {
	template, err := template.New(configuration.filenamePattern).Parse(configuration.filenamePattern)
	if err != nil {
		log.Fatalf("Template couldn't be parsed", err)
	}
	var filenameBuffer = &bytes.Buffer{}
	template.Execute(filenameBuffer, templateData)
	return string(filenameBuffer.Bytes())
}