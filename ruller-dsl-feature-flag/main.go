// seed = {{ index . "_config "seed" }}
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
	mapidc         = 0
	conditionDebug = false
)

func main() {
	logrus.Infof("Starting Ruller DSL Feature Flag code generator")

	logLevel := flag.String("log-level", "info", "debug, info, warning or error")
	source := flag.String("source", "/opt/rules.json", "Comma separated list of files to be used as input json")
	target := flag.String("target", "/opt/rules.go", "Output file name that will be created with the generated Go code")
	condDebug := flag.Bool("condition-debug", false, "Whetever show output nodes with condition info for debugging")
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

	conditionDebug = *condDebug

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

	//PREPARE MAP FOR EACH DSL
	templateRulesMap := make(map[string]interface{})
	for ruleGroupName, v := range jsonRulesMap {
		logrus.Debugf("PROCESSING RULE GROUP %s", ruleGroupName)
		jsonRules := v.(map[string]interface{})
		templateRule := make(map[string]interface{})
		templateRulesMap[ruleGroupName] = templateRule

		//PREPARE CONFIGURATIONS
		logrus.Debugf("CONFIGURATIONS")
		hashSeed := 1234
		flatten := false
		keepFirst := true
		inputTypes := make(map[string]ruller.InputType)
		defaultConditionStr := "true"
		config, exists := jsonRules["_config"].(map[string]interface{})
		if exists {
			dc, exists := config["default_condition"]
			if exists {
				if reflect.ValueOf(dc).Kind() == reflect.String {
					defaultConditionStr = dc.(string)
				} else if reflect.ValueOf(dc).Kind() == reflect.Bool {
					defaultConditionStr = fmt.Sprintf("%t", dc.(bool))
				} else {
					panic(fmt.Errorf("_config default_condition exists but is neither Bool or String type"))
				}
			}

			hs, exists := config["seed"]
			if exists {
				if reflect.ValueOf(hs).Kind() == reflect.Float64 {
					hashSeed = int(hs.(float64))
				} else {
					panic(fmt.Errorf("_config seed exists but is not Float64"))
				}
			}

			ft, exists := config["flatten"]
			if exists {
				if reflect.ValueOf(ft).Kind() == reflect.Bool {
					flatten = ft.(bool)
				} else {
					panic(fmt.Errorf("flatten exists but is not boolean"))
				}
			}

			kf, exists := config["keep_first"]
			if exists {
				if reflect.ValueOf(kf).Kind() == reflect.Bool {
					keepFirst = kf.(bool)
				} else {
					panic(fmt.Errorf("keep_first exists but is not boolean"))
				}
			}

		} else {
			config = make(map[string]interface{})
		}
		config["flatten"] = flatten
		config["keep_first"] = keepFirst
		templateRule["_config"] = config

		//PREPARE "_condition" ATTRIBUTES (generate Go code)
		logrus.Debugf("_CONDITION ATTRIBUTES")
		err := traverseConditionCode(jsonRules, defaultConditionStr, inputTypes, ruleGroupName, fmt.Sprintf("%d", hashSeed))
		if err != nil {
			panic(err)
		}

		// jsonRules["_inputTypes"] = inputTypes
		templateRule["_ruleGroupName"] = ruleGroupName

		//PREPARE GROUP DEFINITIONS
		logrus.Debugf("GROUPS")
		groupCodes := make(map[string]string)
		groups, exists := jsonRules["_groups"].(map[string]interface{})
		if exists {
			//FIXME NEEDED?
			// delete(groups, "_condition")
			for gn, gv := range groups {
				if strings.HasPrefix(gn, "_") {
					continue
				}
				logrus.Debugf(">>>>GROUP %s %s", gn, gv)
				t := reflect.TypeOf(gv)
				if t.Kind() == reflect.Slice {
					garray := ""
					for _, v := range gv.([]interface{}) {
						garray = garray + fmt.Sprintf("\"%s\",", v)
					}
					garray = strings.Trim(garray, ",")
					groupCodes[gn] = fmt.Sprintf("loadGroupArray(groups, \"%s\", \"%s\", []string{%s})", ruleGroupName, gn, garray)

				} else if reflect.ValueOf(gv).Kind() == reflect.String {
					// loadGroupFromFile(groups, "hugeids", "/opt/group1.txt")
					groupCodes[gn] = fmt.Sprintf("loadGroupFromFile(groups, \"%s\", \"%s\", \"%s\")", ruleGroupName, gn, gv.(string))

				} else {
					panic(fmt.Errorf("_groups %s exists but is neither an array of strings nor a string with a file path. rule group %s", gn, ruleGroupName))
				}
			}
		} else {
			logrus.Debugf("No groups found")
		}
		templateRule["_groupCodes"] = groupCodes

		logrus.Debugf("REQUIRED INPUTS")
		requiredInputCodes := make(map[string]string)
		for in, it := range inputTypes {
			icode := fmt.Sprintf("ruller.AddRequiredInput(\"%s\", \"%s\", ruller.%s)", ruleGroupName, in, typeName(it))
			requiredInputCodes[in] = icode
		}
		templateRule["_requiredInputCodes"] = requiredInputCodes

		//ORDERED RULES
		logrus.Debugf("ORDERED RULES")
		rules := make([]map[string]interface{}, 0)
		orderedRules(jsonRules, -1, ruleGroupName, &rules)
		templateRule["_orderedRules"] = rules

		logrus.Debugf("templateRule %s", templateRule)
	}

	logrus.Debugf("Generating Go code")
	sourceCode, err := executeTemplate("/opt/templates", "main.tmpl", templateRulesMap)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("Write Go code to disk")
	err = ioutil.WriteFile(*target, []byte(sourceCode), 0644)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("Code generation finished")
}

func executeTemplate(dir string, templ string, input map[string]interface{}) (string, error) {
	tmpl := template.New("root").Funcs(template.FuncMap{
		"hasPrefix": func(str string, prefix string) bool {
			return strings.HasPrefix(str, prefix)
		},
		"attributeCode": staticAttributeCode,
	})
	tmpl1, err := tmpl.ParseGlob(dir + "/*.tmpl")
	buf := new(bytes.Buffer)
	err = tmpl1.ExecuteTemplate(buf, templ, input)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func staticAttributeCode(attributeName string, attributeValue interface{}, depth int) string {
	result := ""
	mapvar := "output"
	if depth > 0 {
		mapvar = fmt.Sprintf("output%d", depth)
	}

	if reflect.ValueOf(attributeValue).Kind() == reflect.Map {
		if attributeName == "_items" || !strings.HasPrefix(attributeName, "_") {
			map1 := attributeValue.(map[string]interface{})
			nextmapvar := fmt.Sprintf("output%d", depth+1)
			result = result + fmt.Sprintf("%s := make(map[string]interface{})\n			", nextmapvar)
			result = result + fmt.Sprintf("%s[\"%s\"] = %s\n			", mapvar, attributeName, nextmapvar)
			for k, v := range map1 {
				s := staticAttributeCode(k, v, depth+1)
				result = result + s
			}
		}
	} else {
		if !strings.HasPrefix(attributeName, "_") {
			if reflect.ValueOf(attributeValue).Kind() == reflect.Bool {
				result = fmt.Sprintf("%s[\"%s\"] = %t\n			", mapvar, attributeName, attributeValue)
			} else if reflect.ValueOf(attributeValue).Kind() == reflect.Float64 {
				result = fmt.Sprintf("%s[\"%s\"] = %f\n			", mapvar, attributeName, attributeValue)
			} else {
				result = fmt.Sprintf("%s[\"%s\"] = \"%s\"\n			", mapvar, attributeName, attributeValue)
			}
		} else if attributeName == "_condition" && conditionDebug {
			result = fmt.Sprintf("%s[\"%s_debug\"] = \"%s\"\n			", mapvar, attributeName, attributeValue)
		}
	}
	return result
}

func traverseConditionCode(map1 map[string]interface{}, defaultConditionStr string, inputTypes map[string]ruller.InputType, ruleGroupName string, seed string) error {
	createDefaultCondition := true
	for k, v := range map1 {
		// logrus.Debugf("KKKKKK %s %s", k, v)
		if reflect.ValueOf(v).Kind() == reflect.Slice {
			items := v.([]interface{})
			for _, i := range items {
				if reflect.ValueOf(i).Kind() == reflect.Map {
					rm := i.(map[string]interface{})
					traverseConditionCode(rm, defaultConditionStr, inputTypes, ruleGroupName, seed)
				}
			}
		} else if reflect.ValueOf(v).Kind() == reflect.Map {
			if k == "_items" {
				logrus.Debugf("Traversing condition for %s with child items", k)
				traverseConditionCode(v.(map[string]interface{}), defaultConditionStr, inputTypes, ruleGroupName, seed)
			}
		} else {
			if k == "_condition" {
				createDefaultCondition = false
				conditionStr := map1[k]
				map1["_conditionCode"] = conditionCode(conditionStr, inputTypes, ruleGroupName, seed)
			}
		}
	}
	if createDefaultCondition {
		map1["_conditionCode"] = conditionCode(defaultConditionStr, inputTypes, ruleGroupName, seed)
	}
	return nil
}

func orderedRules(map1 map[string]interface{}, parentid int, ruleGroupName string, rules *[]map[string]interface{}) error {
	logrus.Debugf("orderedRules parentid=%d", parentid)
	mapidc = mapidc + 1
	mapid := mapidc
	map1["_id"] = mapid
	map1["_parentid"] = fmt.Sprintf("%d", parentid)
	map1["_ruleGroupName"] = ruleGroupName
	*rules = append(*rules, map1)
	logrus.Debugf("Adding rule %s", map1)
	for k, v := range map1 {
		if k == "_items" {
			logrus.Debugf("attribute %s has children rules", k)
			if reflect.ValueOf(v).Kind() == reflect.Slice {
				logrus.Debugf("attribute %s is an array", k)
				items := v.([]interface{})
				for _, i := range items {
					if reflect.ValueOf(i).Kind() == reflect.Map {
						rm := i.(map[string]interface{})
						logrus.Debugf("attribute %s is an array of maps. calling recursive for item %s", k, i)
						orderedRules(rm, mapid, ruleGroupName, rules)
					}
				}
			} else if reflect.ValueOf(v).Kind() == reflect.Map {
				logrus.Debugf("attribute %s is map. calling recursive", k)
				orderedRules(v.(map[string]interface{}), mapid, ruleGroupName, rules)
			}
			// } else if !strings.HasPrefix(k, "_") {
			// 	logrus.Debugf("attribute %s is a static rule member", k)
		}
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

func conditionCode(value interface{}, inputTypes map[string]ruller.InputType, ruleGroupName string, seed string) string {
	if reflect.ValueOf(value).Kind() == reflect.String {
		condition := value.(string)
		// logrus.Debugf("CONDITION %s", condition)

		//REGEX FUNC
		regexExpreRegex := regexp.MustCompile("(input:[a-z0-9-_]+)\\s*~=\\s*'(.+)'")
		condition = regexExpreRegex.ReplaceAllString(condition, "match($1,\"$2\")")

		//GROUP REFERENCES TO STRING
		//_condition="group:members" ---> ""members""
		groupRegex := regexp.MustCompile("contains\\(\\s*group:([a-z0-9_-]+)\\s*,\\s*([0-9a-z:]+)\\s*\\)")
		condition = groupRegex.ReplaceAllString(condition, "groupContains(\""+ruleGroupName+"\",\"$1\",$2)")

		//RANDOM PERC REFERENCES
		percRegex := regexp.MustCompile("randomPerc\\(\\s*([0-9]+)\\s*,\\s*([0-9a-z:]+)\\s*\\)")
		condition = percRegex.ReplaceAllString(condition, "randomPerc($1,$2,"+seed+")")
		perc2Regex := regexp.MustCompile("randomPercRange\\(\\s*([0-9]+),\\s*([0-9]+)\\s*,\\s*([0-9a-z:]+)\\s*\\)")
		condition = perc2Regex.ReplaceAllString(condition, "randomPercRange($1,$2,$3,"+seed+")")

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
					condition = strings.Replace(condition, "input:" + sm, fmt.Sprintf("input:%s.(string)", sm), -1)
					condition = strings.Replace(condition, ".(string).(string)", ".(string)", -1)

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
		logrus.Debugf("CONDITION %s", condition)
		condition = inputNameRegex.ReplaceAllString(condition, "ctx.Input[\"$1\"]")

		//REPLACE OTHER CHARS
		delimRegex := regexp.MustCompile("'([^']*)'")
		condition = delimRegex.ReplaceAllString(condition, "\"$1\"")
		condition = strings.Replace(condition, " and ", " && ", -1)
		condition = strings.Replace(condition, " or ", " || ", -1)

		logrus.Debugf("CONDITION CODE = %s", condition)

		return condition
	} else {
		panic(fmt.Errorf("Invalid non string '_condition' field. '%v'", value))
	}
}
