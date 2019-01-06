package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/flaviostutz/ruller/ruller"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Sample JSON Rules Server====")

	//call rules registration from generated code
	rules()

	ruller.StartServer()
}
