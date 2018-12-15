package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("Generating code")

	logrus.Debugf("prepare input")
	input := make([]map[string]string, 0)
	for i := 0; i < 5000; i++ {
		rule := make(map[string]string, 0)
		rule["name"] = fmt.Sprintf("name-%d", i)
		rule["label"] = fmt.Sprintf("label-%d", rand.Int())
		rule["screen"] = fmt.Sprintf("screen-%d", rand.Int())
		rule["options"] = fmt.Sprintf("options-%d", rand.Int())
		rule["component"] = fmt.Sprintf("component-%d", rand.Int()%10)
		input = append(input, rule)
	}

	logrus.Debugf("run templates")
	sampleCode, err := executeTemplate("/tmp", input)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("write source code")
	err = ioutil.WriteFile("/opt/main.go", []byte(sampleCode), 0644)
	if err != nil {
		panic(err)
	}

	logrus.Debugf("code generation finished")

	// time.Sleep(100000 * time.Millisecond)
}

func executeTemplate(dir string, input []map[string]string) (string, error) {
	tmpl := template.Must(template.ParseGlob(dir + "/*.tmpl"))
	buf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(buf, "rules.tmpl", input)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
