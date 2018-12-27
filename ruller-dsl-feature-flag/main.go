package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.Infof("Starting Ruller DSL Feature Flag code generator")

	logLevel := flag.String("log-level", "info", "debug, info, warning or error")
	source := flag.String("source", "/opt/rules.json", "Comma separated list of files to be used as input json")
	target := flag.String("target", "/opt/rules.go", "Output file name that will be created with the generated Go code")
	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	sf := strings.Split(*source, ",")
	jsonRulesMap := make(map[string]interface{})
	for _, sourceFile := range sf {
		logrus.Infof("Loading json rules %s", sourceFile)
		jsonFile, err := os.Open(sourceFile)
		if err != nil {
			logrus.Errorf("Error loading json file. err=%s", err)
			os.Exit(1)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var jsonRules map[string]interface{}
		json.Unmarshal([]byte(byteValue), &jsonRules)

		nameregex := regexp.MustCompile("\\/([a-z0-9_-]*)\\..*")
		namer := nameregex.FindStringSubmatch(sourceFile)
		if len(namer) > 1 {
			name := namer[1]
			jsonRulesMap[name] = jsonRules
		} else {
			logrus.Warnf("Couldn't find a valid group rule name in file name. filename=%s", sourceFile)
		}
	}

	logrus.Debugf("Generating Go code")
	sourceCode, err := executeTemplate("/opt/templates", "main.tmpl", jsonRulesMap)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("Write Go code to disk")
	err = ioutil.WriteFile(*target, []byte(sourceCode), 0644)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("code generation finished")
}

func executeTemplate(dir string, templ string, input map[string]interface{}) (string, error) {
	tmpl := template.Must(template.ParseGlob(dir + "/*.tmpl"))
	buf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(buf, templ, input)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
