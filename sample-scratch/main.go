package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/flaviostutz/ruller/ruller"
)

var (
	seed = 1234
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Ruller Sample====")

	//OVERALL CONFIG
	hashSeed := 12345
	ruller.SetDefaultFlatten("menu", true)
	ruller.SetDefaultKeepFirst("menu", true)

	//INITIALIZE GROUPS

	//when set inline
	loadGroupArray(groups, "engineersid", []string{"2345", "1234", "3456", "4567"})
	loadGroupArray(groups, "simples", []string{"245"})

	//when set inline in json
	loadGroupFromFile(groups, "hugeids", "/opt/hugeids.txt")
	loadGroupFromFile(groups, "other", "/opt/other.txt")

	//REQUIRED INPUTS
	ruller.AddRequiredInput("menu", "_remote_ip", ruller.String)
	ruller.AddRequiredInput("menu", "age", ruller.Float64)
	ruller.AddRequiredInput("menu", "customerid", ruller.String)
	ruller.AddRequiredInput("menu", "state", ruller.String)
	ruller.AddRequiredInput("menu", "_remote_ip", ruller.String)
	ruller.AddRequiredInput("menu", "app_version", ruller.String)

	//RULE GROUP 1
	err := ruller.Add("menu", "menu1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["label"] = "menu1"
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.AddChild("menu", "menu1.1", "menu1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		// "_condition": "before('2018-12-31 23:32:21')"
		// "_condition": "after('2018-11-31 23:32:21')"
		// "_condition": "input:age > 30"
		// "_condition": "input:state~='DF|RJ'"
		// "_condition": "randomPercRange(10, 50, input:customerid)",
		// "_condition": "randomPerc(30,input:customerid)",
		// "_condition": "input:_ip_city=='Brasília'",
		// "_condition": "versionCheck(input:app_version, '>1.2.3, <=11.2.3')",

		// "_condition ": "input:age > 30 and random(30,input:customerid)",
		condition := ctx.Input["age"].(float64) > 30 && randomPercRange(10, 30, ctx.Input["customerid"].(string), hashSeed)
		if condition {
			output["label"] = "menu1.1"
			output["uri"] = "/menu1/menu1.1"
			output["component"] = "menu1"

			output1 := make(map[string]interface{})
			output["options"] = output1
			output1["type"] = "brace"
			output1["qtty"] = 123

			output2 := make(map[string]interface{})
			output1["advanced"] = output2
			output2["tip1"] = "abc"
			output2["tip2"] = "xyz"
		}
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.AddChild("menu", "menu1.2", "menu1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		condition := (after("2018-11-30T23:32:21+00:00") && before("2019-11-29T23:32:21+00:00")) ||
			(match(ctx.Input["state"].(string), "DF|RJ") && ctx.Input["_remote_ip"].(string) != "172.1.2.3") &&
				ctx.Input["_ip_city"].(string) == "Brasília" &&
				versionCheck(ctx.Input["app_version"].(string), ">1.2.3, <=11.2.3") || groupContains("hugeids", ctx.Input["customerid"].(string))
		if condition {
			output["label"] = "menu1.2"
			output["uri"] = "/menu1/menu1.2"
			output["component"] = "menu1"
		}
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	//RULE GROUP 2
	err = ruller.Add("domain", "domain1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["domain"] = "domain1.test.com"
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.AddChild("domain", "domain1.1", "domain1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["domain"] = "domain1.1.test.com"
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	ruller.StartServer()
}
