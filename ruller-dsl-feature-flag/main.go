package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/flaviostutz/ruller/ruller"

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

	//find condition fields, validate and collect them
	numberInputRegex := regexp.MustCompile("input:([-_a-z0-9]+)\\s*[><==]\\s*[0-9]+")
	inputNameRegex := regexp.MustCompile("input:([-_a-z0-9]+)")
	inputNameRegex2 := regexp.MustCompile("input:([-_a-z0-9\\.]+)")
	inputTypes := make(map[string]ruller.InputType)

	err := traverseMap(jsonRulesMap, func(ctxMap map[string]interface{}, fieldName string) error {
		if fieldName == "_condition" {
			v := ctxMap[fieldName]
			if reflect.ValueOf(v).Kind() == reflect.String {
				condition := v.(string)

				//find all numeric comparisons
				numberMatches := numberInputRegex.FindAllStringSubmatch(condition, -1)
				for _, numberMatch := range numberMatches {
					//fields uses comparison, so it needs to be float64
					ma := inputNameRegex.FindStringSubmatch(numberMatch[1])
					attributeName := ma[1]
					logrus.Debugf("Updating attribute '%s' to '%s'", attributeName, fmt.Sprintf("%s.(float64)", attributeName))
					condition = strings.Replace(condition, attributeName, fmt.Sprintf("%s.(float64)", attributeName), -1)

					//check and collect input types
					it, exists := inputTypes[attributeName]
					if exists {
						if it != ruller.Float64 {
							panic(fmt.Errorf("Attribute '%s' was defined as '%v' and now is being redefined as Float64. Aborting", attributeName, it))
						}
					} else {
						inputTypes[attributeName] = ruller.Float64
					}
				}

				//cast all other attributes to string
				matches := inputNameRegex2.FindAllStringSubmatch(condition, -1)
				for _, match := range matches {
					if len(match) > 1 {
						sm := match[1]
						//update all matches that hasn't been changed on previous step
						if !strings.Contains(sm, ".") {
							logrus.Debugf("Updating attribute '%s' to '%s'", sm, fmt.Sprintf("%s.(string)", sm))
							condition = strings.Replace(condition, sm, fmt.Sprintf("%s.(string)", sm), -1)
						} else {
							logrus.Debugf("Ignoring already casted attribute %s", sm)
						}
						//check and collect input types
						it, exists := inputTypes[sm]
						if exists {
							if it != ruller.Float64 {
								panic(fmt.Errorf("Attribute '%s' was defined as '%v' and now is being redefined as Float64. Aborting", sm, it))
							}
						} else {
							inputTypes[sm] = ruller.String
						}
					}
				}
				ctxMap[fieldName] = condition
			} else {
				panic(fmt.Errorf("Invalid non string '_condition' field. '%v'", v))
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	logrus.Debugf("Generating Go code")
	sourceCode, err := executeTemplate("/opt/templates", "main.tmpl", jsonRulesMap)
	logrus.Debugf("SOURCE CODE:\\n%s", sourceCode)
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

//FieldProcessor called in traverseMap
type FieldProcessor func(contextMap map[string]interface{}, fieldName string) error

func traverseMap(map1 map[string]interface{}, processor FieldProcessor) error {
	for k, v := range map1 {
		if reflect.ValueOf(v).Kind() == reflect.Map {
			traverseMap(v.(map[string]interface{}), processor)
		} else {
			err := processor(map1, k)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
