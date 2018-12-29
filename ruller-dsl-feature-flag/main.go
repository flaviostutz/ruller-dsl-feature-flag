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

var (
	defaultConditionStr = "true"
	inputTypes          = make(map[string]ruller.InputType)
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

	logrus.Debugf("json rules menu %s", jsonRulesMap["menu"])
	logrus.Debugf("json rules domains %s", jsonRulesMap["domains"])

	//CREATE CODE FOR EACH DSL
	sourceCode := ""
	for name, v := range jsonRulesMap {
		jsonRules := v.(map[string]interface{})
		//configurations
		hashSeed := 1234
		config, exists := jsonRules["_config"].(map[string]interface{})
		if exists {
			dc, exists := config["default_condition"]
			if exists {
				if reflect.ValueOf(dc).Kind() == reflect.String {
					defaultConditionStr = dc.(string)
				} else if reflect.ValueOf(dc).Kind() == reflect.Bool {
					defaultConditionStr = fmt.Sprintf("%t", dc.(bool))
				} else {
					panic(fmt.Errorf("default_condition exists but is neither Bool or String type"))
				}
			}

			hs, exists := config["seed"]
			if exists {
				if reflect.ValueOf(dc).Kind() == reflect.Float64 {
					hashSeed = int(hs.(float64))
				} else {
					panic(fmt.Errorf("default_condition exists but is not Float64"))
				}
			}
		} else {
			config = make(map[string]interface{})
			jsonRules["_config"] = config
		}
		config["seed"] = hashSeed

		//convert "_condition" attribute to Go code
		err := traverseRulesMap(jsonRules, func(ctxMap map[string]interface{}, fieldName string) error {
			// logrus.Debugf("TRAVERSE %s", fieldName)
			if fieldName == "_condition" {
				conditionStr := ctxMap[fieldName]
				ctxMap[fieldName] = conditionCode(conditionStr, inputTypes)
			}
			return nil
		}, defaultConditionStr)
		if err != nil {
			panic(err)
		}

		logrus.Debugf("MAAAAAP %s", jsonRules)

		logrus.Debugf("Generating Go code")
		sourceCode, err = executeTemplate("/opt/templates", "main.tmpl", jsonRules)
		if err != nil {
			panic(err)
		}
	}

	logrus.Debugf("Write Go code to disk")
	err := ioutil.WriteFile(*target, []byte(sourceCode), 0644)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("Code generation finished")
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

func traverseRulesMap(map1 map[string]interface{}, processor FieldProcessor, defaultConditionStr string) error {
	// logrus.Debugf("MMMMMMMMMM %s", map1)
	conditionFound := false
	for k, v := range map1 {
		// logrus.Debugf("KKKKKK %s %s", k, v)
		if reflect.ValueOf(v).Kind() == reflect.Slice {
			items := v.([]interface{})
			for _, i := range items {
				if reflect.ValueOf(i).Kind() == reflect.Map {
					rm := i.(map[string]interface{})
					traverseRulesMap(rm, processor, defaultConditionStr)
				}
			}
		} else if reflect.ValueOf(v).Kind() == reflect.Map {
			traverseRulesMap(v.(map[string]interface{}), processor, defaultConditionStr)
		} else {
			if k == "_condition" {
				conditionFound = true
			}
			err := processor(map1, k)
			if err != nil {
				return err
			}
		}
	}
	if !conditionFound {
		map1["_condition"] = conditionCode(defaultConditionStr, inputTypes)
	}
	return nil
}

func typeName(inputType ruller.InputType) string {
	if inputType == ruller.String {
		return "String"
	} else if inputType == ruller.Float64 {
		return "Float64"
	} else if inputType == ruller.Bool {
		return "Bool"
	} else {
		return "-"
	}
}

func conditionCode(value interface{}, inputTypes map[string]ruller.InputType) string {
	if reflect.ValueOf(value).Kind() == reflect.String {
		condition := value.(string)
		// logrus.Debugf("CONDITION %s", condition)

		//REGEX FUNC
		regexExpreRegex := regexp.MustCompile("(input:[a-z0-9-_]+)\\s*~=\\s*'(.+)'")
		condition = regexExpreRegex.ReplaceAllString(condition, "match($1,\"$2\")")

		//ADD CASTS
		//_condition="input:age > 30 and input:name='stutz'" ---> "input:age.(float64) > 30 and input:name.(string)=='stutz'"

		//find all numeric comparisons
		numberInputRegex := regexp.MustCompile("input:([a-z0-9-_]+)\\s*[><==]\\s*[0-9]+")
		numberMatches := numberInputRegex.FindAllStringSubmatch(condition, -1)
		for _, numberMatch := range numberMatches {
			logrus.Debugf("Condition number match %s - %s", numberMatch[0], numberMatch[1])
			//fields uses comparison, so it needs to be float64
			attributeName := numberMatch[1]
			logrus.Debugf("Updating attribute '%s' to '%s'", attributeName, fmt.Sprintf("%s.(float64)", attributeName))
			condition = strings.Replace(condition, attributeName, fmt.Sprintf("%s.(float64)", attributeName), -1)

			//check and collect input types
			it, exists := inputTypes[attributeName]
			if exists {
				if it != ruller.Float64 {
					panic(fmt.Errorf("Attribute '%s' was defined as '%v' and now is being redefined as 'Float64'. Aborting", attributeName, typeName(it)))
				}
			} else {
				inputTypes[attributeName] = ruller.Float64
				logrus.Debugf("Input %s is Float64", attributeName)
			}
		}

		//cast all other attributes to string
		inputNameRegex2 := regexp.MustCompile("input:([a-z0-9-_\\.]+)")
		matches := inputNameRegex2.FindAllStringSubmatch(condition, -1)
		for _, match := range matches {
			if len(match) > 1 {
				sm := match[1]
				//update all matches that hasn't been changed on previous step
				if !strings.Contains(sm, ".") {
					logrus.Debugf("Updating attribute '%s' to '%s'", sm, fmt.Sprintf("%s.(string)", sm))
					condition = strings.Replace(condition, sm, fmt.Sprintf("%s.(string)", sm), -1)

					//check and collect input types
					it, exists := inputTypes[sm]
					if exists {
						if it != ruller.String {
							panic(fmt.Errorf("Attribute '%s' was defined as '%v' and now is being redefined as 'String'. Aborting", sm, typeName(it)))
						}
					} else {
						inputTypes[sm] = ruller.String
						logrus.Debugf("Input %s is String", sm)
					}
				} else {
					logrus.Debugf("Ignoring already casted attribute %s", sm)
				}
			}
		}

		//GET INPUT FROM CONTEXT
		//_condition="input:age > 30 and input:name='stutz'" ---> "ctx.Input["age"].(float64) > 30 and ctx.Input["name"].(string)=="stutz""
		inputNameRegex := regexp.MustCompile("input:([a-z0-9-_]+)")
		condition = inputNameRegex.ReplaceAllString(condition, "ctx.Input[\"$1\"]")

		//GROUP REFERENCES TO STRING
		//_condition="group:members" ---> ""members""
		groupRegex := regexp.MustCompile("group:([a-z0-9-_]+)")
		condition = groupRegex.ReplaceAllString(condition, "\"$1\"")

		//REPLACE OTHER CHARS
		delimRegex := regexp.MustCompile("'(.*)'")
		condition = delimRegex.ReplaceAllString(condition, "\"$1\"")
		condition = strings.Replace(condition, " and ", " && ", -1)
		condition = strings.Replace(condition, " or ", " || ", -1)

		logrus.Debugf("CONDITION CODE=%s", condition)

		return condition
	} else {
		panic(fmt.Errorf("Invalid non string '_condition' field. '%v'", value))
	}
}
